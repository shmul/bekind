package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/phuslu/log"
	"github.com/robfig/cron/v3"
	"github.com/shmul/bekind/pkg/dns"
	"github.com/shmul/bekind/pkg/gcp"
)

type (
	version struct{}
	dnsOpts struct {
		Domain        string            `short:"d" long:"domain" description:"base domain to use" required:"true"`
		Port          int               `short:"p" long:"port" description:"Bind port" default:"7353"`
		TTL           int               `long:"ttl" description:"default TTL (secs)" default:"30"`
		MappedRecords map[string]string `short:"r" long:"record" description:"mapped dns record - name:address"`
		SelfRecords   []string          `short:"s" long:"self" description:"self dns record (own IP) - name"`
		ExternalIP    string            `long:"self-addr" description:"self IP address"`
	}
)

const (
	DefaultConfigFile = "config.ini"
)

var (
	Branch    string
	Timestamp string
	Revision  string

	opts struct {
		Verbose    bool    `short:"v" long:"verbose" description:"Show verbose debug information"`
		Level      string  `long:"level" default:"info" description:"log level" value-name:"LEVEL"`
		StdoutOnly bool    `long:"stdout" description:"log only to stdout when running in terminal"`
		Version    version `command:"version"`
		DNS        dnsOpts `command:"dns"`
	}
	parser = flags.NewParser(&opts, flags.Default)
)

func setupLogging() {
	log.DefaultLogger = log.Logger{
		Level:      log.ParseLevel(opts.Level),
		TimeFormat: "06-01-02T15:04:05.999",
		Caller:     0,
	}

	var terminalWriter log.Writer
	if log.IsTerminal(os.Stderr.Fd()) {
		terminalWriter = &log.ConsoleWriter{
			ColorOutput:    true,
			QuoteString:    true,
			EndWithMessage: true,
		}
	}

	if opts.StdoutOnly {
		log.DefaultLogger.Writer = terminalWriter
		return
	}

	fileWriter := &log.FileWriter{
		Filename:     "logs/bekind.log",
		FileMode:     0600,
		MaxSize:      100 * 1024 * 1024,
		MaxBackups:   7,
		EnsureFolder: true,
	}

	runner := cron.New(cron.WithLocation(time.Local))
	runner.AddFunc("0 0 * * *", func() { fileWriter.Rotate() })
	go runner.Run()

	if terminalWriter != nil {
		log.DefaultLogger.Writer = &log.MultiEntryWriter{
			terminalWriter,
			fileWriter,
		}
	} else {
		log.DefaultLogger.Writer = fileWriter
	}
}

func setupAndExecute(command flags.Commander, args []string) error {
	setupLogging()

	if opts.Verbose {
		flags.NewIniParser(parser).Write(os.Stdout, flags.IniDefault)
	}
	return command.Execute(args)
}

func main() {
	parser.CommandHandler = setupAndExecute
	f := os.Getenv("BEKIND_CONFIG_FILE")
	if f == "" {
		f = DefaultConfigFile
	}
	_, err := os.Stat(f)
	if err == nil {
		err = flags.NewIniParser(parser).ParseFile(f)
		if err != nil {
			log.Error().Err(err)
			return
		}
	}
	_, err = parser.Parse()

	if err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			log.Error().Err(err)
			os.Exit(1)
		default:
			log.Error().Err(err)
			os.Exit(1)
		}
	}
}

func (v *version) Execute(args []string) error {
	fmt.Println("Branch:    ", Branch)
	fmt.Println("Revision:  ", Revision)
	fmt.Println("Timestamp: ", Timestamp)
	return nil
}

func (d *dnsOpts) Execute(args []string) error {
	records := make(map[string]net.IP)
	for k, v := range d.MappedRecords {
		records[k] = net.ParseIP(v)
	}
	selfAddr, err := gcp.ExternalIP()
	if err != nil {
		log.Warn().Err(err).Msg("Execute")
		selfAddr = net.ParseIP(d.ExternalIP)
	}
	if selfAddr != nil {
		log.Info().IPAddr("self-addr", selfAddr).Msg("Execute")
	}
	for _, v := range d.SelfRecords {
		records[v] = selfAddr
	}

	server, err := dns.New(dns.Config{Port: d.Port, Domain: d.Domain, Fixed: records})
	if err != nil {
		log.Fatal().Err(err)
	}
	return server.Run()
}
