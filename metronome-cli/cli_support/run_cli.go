package cli

import (
	"flag"
	"io"
	"fmt"
	"bytes"
	"errors"
	log "github.com/behance/go-logrus"
	"os"
)

// RunsTopLevel - top level cli menuing structure
//  Implements CommandParse and CommandExecute
//  Parses actions without flags package
type RunsTopLevel JobTopLevel

// Usage - CommandParse implementation
func (theRun *RunsTopLevel) Usage(writer io.Writer) {
	fmt.Fprintf(writer, "run {start|stop|ls|get} <options>:\n")
	fmt.Fprintln(writer, `
	  start <options>  | Start a Job that has a schedule.
	  stop  <options>  | Stop a Job
	  ls               | Status a Job -- currently only returns 'ACTIVE' jobs
	  get <options>    | Get a Job run status.

	  Call run <action> help for more on a sub-command
	`)
}
// Parse - parse the top level `run <action>` menu
//
func (theRun *RunsTopLevel) Parse(args [] string) (exec CommandExec, err error) {

	defer func() {
		if r := recover(); r != nil {
			buf := new(bytes.Buffer)
			fmt.Fprintln(buf, r.(error).Error())
			fmt.Fprintf(buf, "\n %s usage:\n", theRun.subcommand)

			if theRun.task != nil {
				theRun.task.Usage(buf)
			}

			theRun.Usage(buf)
			err = errors.New(buf.String())
		}
	}()
	if len(args) == 0 {
		panic(errors.New("sub command required"))
	}
	log.Debugf("RunTopLevel args: %+v", args)
	theRun.subcommand = args[0]
	switch theRun.subcommand {
	case "ls":
		theRun.task = CommandParse(new(RunLs))
	case "get":
		theRun.task = CommandParse(new(RunStatusJob))
	case "start":
		theRun.task = CommandParse(new(RunStartJob))
	case "stop":
		theRun.task = CommandParse(new(RunStopJob))
	case "help", "--help":
		theRun.Usage(os.Stderr)
		return nil, errors.New("run usage")
	default:
		return nil, fmt.Errorf("run: unknown action '%s'", theRun.subcommand)
	}
	var subcommandArgs []string
	if len(args) > 1 {
		subcommandArgs = args[1:]
	}
	log.Debugf("run %s args: %+v", theRun.subcommand, subcommandArgs)
	if exec, err = theRun.task.Parse(subcommandArgs); err != nil {
		panic(err)
	} else {
		return exec, err
	}
}

// RunLs - Get all the runs for a job via cli
// GET /v1/jobs/$jobId/runs
type RunLs JobID

// Usage - RunLs usage
func (theRun *RunLs) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("job ls", flag.ExitOnError)
	(*JobID)(theRun).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}
// Parse - parse flags
//   - need a job-id
//   - implements CommandParse
func (theRun *RunLs) Parse(args []string) (_ CommandExec, err error) {
	flags := flag.NewFlagSet("run ls", flag.ExitOnError)
	(*JobID)(theRun).FlagSet(flags)
	defer func() {
		if r := recover(); r != nil {
			buf := new(bytes.Buffer)
			flags.SetOutput(buf)
			fmt.Fprintln(buf, err.Error())
			err = errors.New(buf.String())
		}
	}()
	if err = flags.Parse(args); err != nil {
		panic(err)
	} else if err = (*JobID)(theRun).Validate(); err != nil {
		panic(err)
	} else {
		return theRun, nil
	}
}
// Execute the Metronome API
func (theRun *RunLs) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.Runs(string(*theRun),0)
}

// RunStartJob - cli actuator to run POST /v1/jobs/$jobId/runs
//   - Really same as JobID but uses different flags
type RunStartJob JobID

// Usage - Start the job usage
func (theRun *RunStartJob) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("run start", flag.ExitOnError)
	(*JobID)(theRun).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}
// Parse - Parse the flags
func (theRun *RunStartJob) Parse(args []string) (_ CommandExec, err error) {
	flags := flag.NewFlagSet("run start", flag.ExitOnError)
	(*JobID)(theRun).FlagSet(flags)
	defer func() {
		if r := recover(); r != nil {
			buf := new(bytes.Buffer)
			flags.SetOutput(buf)
			fmt.Fprintln(buf, err.Error())
			err = errors.New(buf.String())
		}
	}()
	if err = flags.Parse(args); err != nil {
		panic(err)
	} else if err = (*JobID)(theRun).Validate(); err != nil {
		panic(err)
	} else {
		return theRun, nil
	}
}
// Execute - the api against Metronome
func (theRun *RunStartJob) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.StartJob(string(*theRun))
}

// RunStatusJob - cli actuator that runs `GET  /v1/jobs/$jobId/runs/$runId`
type RunStatusJob struct {
	JobID
	RunID
}
// FlagSet - set up flags to get the status
func (theRun *RunStatusJob) FlagSet(flags *flag.FlagSet) *flag.FlagSet {
	theRun.JobID.FlagSet(flags)
	theRun.RunID.FlagSet(flags)
	return flags
}
// Validate - validate the flags
func (theRun *RunStatusJob) Validate() error {
	if err := theRun.JobID.Validate(); err != nil {
		return err
	} else if err = theRun.RunID.Validate(); err != nil {
		return err
	}
	return nil
}
// Usage - run status
func (theRun *RunStatusJob) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("run status", flag.ExitOnError)
	theRun.FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}
// Parse - run-id,job-id
// 	- status flags.
//   	- panics returned as error
//	- implements CommandParse
//
func (theRun *RunStatusJob) Parse(args []string) (_ CommandExec, err error) {
	flags := flag.NewFlagSet("run status", flag.ExitOnError)
	theRun.FlagSet(flags)

	defer func() {
		if r := recover(); r != nil {
			buf := new(bytes.Buffer)
			flags.SetOutput(buf)
			fmt.Fprintln(buf, err.Error())
			err = errors.New(buf.String())
		}
	}()
	if err = flags.Parse(args); err != nil {
		panic(err)
	} else if err = theRun.Validate(); err != nil {
		panic(err)
	} else {
		return theRun, nil
	}
}
// Execute - executes the job status against metronome
func (theRun *RunStatusJob) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.StatusJob(string(theRun.JobID), string(theRun.RunID))
}

// RunStopJob - cli structure facilitating POST /v1/jobs/$jobId/runs/$runId/action/stop
//  - implements both CommandParse and CommandExecute interfaces
//  - light-weight override of run status cli structure (same args,flags)
//  - used via cli to capture, execute:
type RunStopJob RunStatusJob

// Usage - implementation of cli usage needed to RunStatusJob
func (theRun *RunStopJob) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("run stop", flag.ExitOnError)
	(*RunStatusJob)(theRun).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}
// Parse - takes cli arguments and assures valid flags are passed to get stop a job (job-id,run-id)
//   - uses RunStatusJob to process flags but uses itself as the CommandExecute implementation
//   - Implements CommandParse & CommandExecut
//   - uses runstatus flags but returns itself as the CommandExecute implementation
func (theRun *RunStopJob) Parse(args []string) (_ CommandExec, err error) {
	flags := flag.NewFlagSet("run stop", flag.ExitOnError)
	(*RunStatusJob)(theRun).FlagSet(flags)
	defer func() {
		if r := recover(); r != nil {
			buf := new(bytes.Buffer)
			flags.SetOutput(buf)
			fmt.Fprintln(buf, err.Error())
			err = errors.New(buf.String())
		}
	}()

	if err = flags.Parse(args); err != nil {
		panic(err)
	} else if err = (*RunStatusJob)(theRun).Validate(); err != nil {
		panic(err)
	}
	return theRun, nil
}
// Execute - executes POST /v1/jobs/$jobId/runs/$runId/action/stop
func (theRun *RunStopJob) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.StopJob(string(theRun.JobID), string(theRun.RunID))
}
