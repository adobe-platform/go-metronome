package cli

import (
	"flag"
	met "github.com/adobe-platform/go-metronome/metronome"
	"github.com/Sirupsen/logrus"
	"io"
)
//Runtime represents the global options passed to all CommandExec.Execute methods.
//In particular, it keeps the metronome client and the other useful global options
type Runtime struct {
	httpAddr string
	flags    *flag.FlagSet
	Debug bool
	help     bool
	client   met.Metronome
}

//
// Global flags are kept in 'Runtime'.  main takes care of sending Parse the correct list of args
//
func (self *Runtime) FlagSet(name  string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	flags.StringVar(&self.httpAddr, "metronome-url", DefaultHTTPAddr, "Set the Metronome address")
	flags.BoolVar(&self.Debug, "debug", false, "Turn on debug")
	return flags
}
func (self *Runtime) Usage(writer io.Writer) {
	flags := self.FlagSet("<global options help>")
	flags.SetOutput(writer)
	flags.PrintDefaults()
}
func (self *Runtime) Parse(args []string) (CommandExec, error) {
	flags := self.FlagSet("<global options> ")
	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	config := met.NewDefaultConfig()
	config.URL = self.httpAddr
	if client, err := met.NewClient(config); err != nil {
		return nil, err
	} else {
		self.client = client
	}
	logrus.Debugf("Runtime <global flags> ok\n")
	// No exec returned
	return nil, nil
}
