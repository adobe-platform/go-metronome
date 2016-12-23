package cli

import (
	"flag"
	met "github.com/adobe-platform/go-metronome/metronome"
	"github.com/Sirupsen/logrus"
	"io"
	"fmt"
	"strings"
)

//Runtime represents the global options passed to all CommandExec.Execute methods.
//In particular, it keeps the metronome client and the other useful global options
type Runtime struct {
	httpAddr  string
	flags     *flag.FlagSet
	Debug     bool
	help      bool
	client    met.Metronome
	authToken string
	user      string
	pw        string
}

//
// Global flags are kept in 'Runtime'.  main takes care of sending Parse the correct list of args
//

// FlagSet - Set up the flags
func (runtime *Runtime) FlagSet(name  string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	flags.StringVar(&runtime.httpAddr, "metronome-url", DefaultHTTPAddr, "Set the Metronome address")
	flags.BoolVar(&runtime.Debug, "debug", false, "Turn on debug")
	flags.StringVar(&runtime.authToken, "authorization", "", "Authorization token")
	flags.StringVar(&runtime.user, "user", "", "user")
	flags.StringVar(&runtime.pw, "password", "", "password")
	return flags
}
// Usage - emit the usage
func (runtime *Runtime) Usage(writer io.Writer) {
	flags := runtime.FlagSet("<global options help>")
	flags.SetOutput(writer)
	flags.PrintDefaults()
}
// Parse - Process command line arguments
func (runtime *Runtime) Parse(args []string) (CommandExec, error) {
	flags := runtime.FlagSet("<global options> ")
	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	config := met.NewDefaultConfig()
	config.URL = runtime.httpAddr
	if runtime.authToken != "" {
		if strings.Contains(runtime.authToken, "token=") {
			config.AuthToken = runtime.authToken
		} else {
			config.AuthToken = fmt.Sprintf("token=%s", runtime.authToken)
		}
	}
	if runtime.user != "" {
		config.User = runtime.user
	}
	if runtime.pw != "" {
		config.Pw = runtime.pw
	}
	if runtime.Debug {
		config.Debug = runtime.Debug
	}

	client, err := met.NewClient(config)
	if err != nil {
		return nil, err
	}
	runtime.client = client

	logrus.Debugf("Runtime <global flags> ok")
	// No exec returned
	return nil, nil
}
