package cli

import (
	log "github.com/behance/go-logrus"
	"fmt"
	"io"

)

// Metrics top level
//  GET  /v1/metrics
type Metrics int
// Usage - emit usage instructions
func (metrics *Metrics) Usage(writer io.Writer) {
	fmt.Fprintf(writer, "\nmetrics  -  dumps metronome metrics\n")
}
// Parse - parse any expected args.  None in the this case but conforming to interface
func (metrics *Metrics) Parse(args []string) (CommandExec, error) {
	return metrics, nil
}
// Execute - run the metrics REST command
func (metrics *Metrics) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.Metrics()
}

// Ping - type implementing Parse/Execute interfaces for
//  GET /v1/ping
type Ping int

// Usage - emit the Ping usage instructions. Implements Parse interface method
func (ping *Ping) Usage(writer io.Writer) {
	fmt.Fprintf(writer, "\nping  - pings metronome\n")
}

// Parse - implement interface notably returning the self as the Executor interface
func (ping *Ping) Parse(args []string) (CommandExec, error) {
	log.Debugf("Ping.Parse: %+v", args)
	return ping, nil
}
// Execute - run the Metronome Ping command against he metronome service
func (ping *Ping) Execute(runtime *Runtime) (interface{}, error) {
	msg, err := runtime.client.Ping()
	if err != nil {
		return nil, err
	}
	return msg, nil
}


