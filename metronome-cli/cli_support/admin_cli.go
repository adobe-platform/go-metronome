package cli

import (
	"github.com/Sirupsen/logrus"
	"fmt"
	"io"

)

// Metrics top level
//  GET  /v1/metrics
type Metrics int

func (self *Metrics) Usage(writer io.Writer) {
	fmt.Fprintf(writer, "dumps metronome metrics\n")
}

func (self *Metrics) Parse(args []string) (CommandExec, error) {
	return self, nil
}
func (self *Metrics) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.Metrics()
}

//  GET /v1/ping
type Ping int

func (self *Ping) Usage(writer io.Writer) {
	fmt.Fprintf(writer, "ping  - pings metronome\n")
}

func (self *Ping) Parse(args []string) (CommandExec, error) {
	logrus.Debugf("Ping.Parse: %+v\n", args)
	return self, nil
}
func (self *Ping) Execute(runtime *Runtime) (interface{}, error) {
	logrus.Debugf("Ping.execute\n")
	if msg, err := runtime.client.Ping(); err != nil {
		return nil, err
	} else {
		return msg, nil
	}
}


