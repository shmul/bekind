package dns

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	miekgdns "github.com/miekg/dns"
	"github.com/phuslu/log"
)

type Config struct {
	Port   int
	Domain string
	Fixed  map[string]net.IP
}

type Server struct {
	Config
	impl *miekgdns.Server
	l    log.Logger
}

func New(c Config) (*Server, error) {
	s := &Server{
		Config: c,
		l:      log.DefaultLogger,
	}
	s.l.Context = log.NewContext(nil).Str("pkg", "dns").Value()
	if !strings.HasSuffix(s.Domain, ".") {
		s.Domain = s.Domain + "."
	}
	s.impl = &miekgdns.Server{Addr: ":" + strconv.Itoa(c.Port), Net: "udp"}
	s.l.Info().Str("domain", s.Domain).Int("port", s.Port).Msg("Server")
	return s, nil
}

func (s *Server) parseQuery(m *miekgdns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case miekgdns.TypeA:
			s.l.Info().Str("q", q.Name).Msg("parseQuery")
			stripped := strings.TrimSuffix(q.Name, "."+s.Domain)
			ip := s.Fixed[stripped]
			if ip != nil {
				rr, err := miekgdns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip.String()))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				} else {
					s.l.Warn().Err(err).Msg("parseQuery")
				}
			}
		}
	}
}

func (s *Server) handle(w miekgdns.ResponseWriter, r *miekgdns.Msg) {
	m := new(miekgdns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case miekgdns.OpcodeQuery:
		s.parseQuery(m)
	}

	w.WriteMsg(m)
}

func (s *Server) Run() error {
	// attach request handler func
	miekgdns.HandleFunc(s.Domain, s.handle)

	s.l.Info().Msg("Run - enter")
	defer s.impl.Shutdown()
	err := s.impl.ListenAndServe()
	if err != nil {
		s.l.Error().Err(err).Msg("Run")
		return err
	}

	s.l.Info().Msg("Run - Done")
	return nil
}
