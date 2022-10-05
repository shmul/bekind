package dns

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	miekgdns "github.com/miekg/dns"
	"github.com/phuslu/log"
	"github.com/shmul/bekind/pkg/ids"
)

type Config struct {
	Port   int
	Domain string
	Fixed  map[string]net.IP
	TTL    int
}

type Server struct {
	Config
	impl *miekgdns.Server
	l    log.Logger
	g    func() string
}

var helpLines = []string{
	"ip, my, myip - caller ip address",
	"id, key, nanoid - random id",
}

func New(c Config) (*Server, error) {
	s := &Server{
		Config: c,
		l:      log.DefaultLogger,
	}
	s.l.Context = log.NewContext(nil).Str("pkg", "dns").Value()
	s.g, _ = ids.Generator(12)
	if !strings.HasSuffix(s.Domain, ".") {
		s.Domain = s.Domain + "."
	}
	s.impl = &miekgdns.Server{Addr: ":" + strconv.Itoa(c.Port), Net: "udp"}
	s.l.Info().Str("domain", s.Domain).Int("port", s.Port).Msg("Server")
	return s, nil
}

func (s *Server) parseQuery(m *miekgdns.Msg, remote net.Addr) {
	answer := func(record string) error {
		rr, err := miekgdns.NewRR(record)
		if err == nil {
			m.Answer = append(m.Answer, rr)
			return nil
		}
		s.l.Warn().Err(err).Msg("parseQuery")
		return err
	}

	txtAnswer := func(name, msg string) error {
		return answer(fmt.Sprintf("%s 1 TXT %s", name, msg))
	}

	aRecord := func(q miekgdns.Question) bool {
		s.l.Info().Str("q", q.Name).Msg("parseQuery")
		stripped := strings.TrimSuffix(q.Name, "."+s.Domain)
		if ip, exists := s.Fixed[stripped]; exists &&
			answer(fmt.Sprintf("%s %d A %s", q.Name, s.TTL, ip.String())) == nil {
			return true
		}
		switch stripped {

		case "ip", "my", "myip":
			s.l.Info().Str("ip", remote.String()).Msg("parseQuery")
			parts := strings.Split(remote.String(), ":")
			if txtAnswer(q.Name, parts[0]) == nil {
				return true
			}

		case "id", "key", "nanoid":
			// TODO - also look for N.id and generate N lengthed ids
			if txtAnswer(q.Name, s.g()) == nil {
				return true
			}
		}

		return false
	}

	for _, q := range m.Question {
		switch q.Qtype {
		case miekgdns.TypeA:
			if aRecord(q) {
				return
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
		s.parseQuery(m, w.RemoteAddr())
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
