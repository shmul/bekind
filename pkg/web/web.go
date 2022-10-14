package web

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/acme"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mcuadros/go-defaults"
	"github.com/phuslu/log"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sys/unix"
	"golang.org/x/time/rate"
)

type Config struct {
	ListenAddrPort string `default:"127.0.0.1:443"`
	RootDir        string
	CacheDir       string
	Hosts          []string
	RateLimit      int `default:"100"`
}

type Web struct {
	c Config
	e *echo.Echo
	s http.Server
	l log.Logger

	ctx context.Context
}

type RouteSetup struct {
	Prefix string
	Setup  func(w *Web, g *echo.Group)
}

func accessible(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return err
	}
	return unix.Access(path, unix.W_OK)
}

func New(ctx context.Context, c Config) (*Web, error) {
	defaults.SetDefaults(&c)
	w := &Web{
		c:   c,
		e:   echo.New(),
		ctx: ctx,
		l:   log.DefaultLogger,
	}
	w.l.Context = log.NewContext(nil).Str("pkg", "web").Value()

	if err := accessible(c.CacheDir); err != nil {
		return nil, err
	}

	if err := accessible(c.RootDir); err != nil {
		return nil, err
	}

	w.e.AutoTLSManager.Cache = autocert.DirCache(c.CacheDir)
	autoTLSManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		// Cache certificates to avoid issues with rate limits (https://letsencrypt.org/docs/rate-limits)
		Cache:      w.e.AutoTLSManager.Cache,
		HostPolicy: autocert.HostWhitelist(w.c.Hosts...),
	}

	w.s = http.Server{
		Addr:    w.c.ListenAddrPort,
		Handler: w.e, // set Echo as handler
		TLSConfig: &tls.Config{
			// Certificates: nil, // <-- s.ListenAndServeTLS will populate this field
			GetCertificate: autoTLSManager.GetCertificate,
			NextProtos:     []string{acme.ALPNProto},
		},
		ReadTimeout: 10 * time.Second, // use custom timeouts
	}

	return w, nil
}

func (w *Web) SetupRoutes(handlers []RouteSetup) error {
	for _, h := range handlers {
		if !strings.HasPrefix(h.Prefix, "/") {
			h.Prefix = "/" + h.Prefix
		}
		h.Setup(w, w.e.Group(h.Prefix))
	}
	return nil
}

func (w *Web) Run() error {
	w.e.Use(
		middleware.Recover(),
		middleware.Logger(),
		middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(w.c.RateLimit))),
	)
	onLocalhost := strings.HasPrefix(w.c.ListenAddrPort, "localhost:") || strings.HasPrefix(w.c.ListenAddrPort, "127.0.0.1:")
	w.l.Info().Bool("localhost", onLocalhost).Msg("Run - start")
	var err error
	if onLocalhost {
		err = w.s.ListenAndServe()
	} else {
		err = w.s.ListenAndServeTLS("", "")
	}
	if err != http.ErrServerClosed {
		w.l.Error().Err(err).Msg("Run")
		return err
	}
	w.l.Info().Msg("Run - end")
	return nil
}
