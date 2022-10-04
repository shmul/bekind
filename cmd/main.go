package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

var (
	Branch    string
	Timestamp string
	Revision  string

	version struct{}
	opts    struct {
		Verbose bool `short:"v" long:"verbose" description:"Show verbose debug information"`

		Version version `command:"version"`

		DNS struct {
			Port int `short:"p" long:"port" description:"Bind port" default:"53"`
		} `command:"dns"`
	}
	parser = flags.NewParser(&opts, flags.Default)
)

func main() {
	if _, err := parser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		default:
			os.Exit(1)
		}
	}
}

func (v *version) Execute(args []string) error {
	fmt.Println("Branch", Branch)
	fmt.Println("Revision", Revision)
	fmt.Println("Timestamp", Timestamp)
}
