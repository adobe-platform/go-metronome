package cli

import (
	"flag"
	"io"
	"fmt"
	"bytes"
	"errors"
	"github.com/Sirupsen/logrus"
	"os"
)

type RunsTopLevel JobTopLevel

func (self *RunsTopLevel) Usage(writer io.Writer) {
	fmt.Fprintf(writer, "run {start|stop|ls|get} <options>:\n")
	fmt.Fprintln(writer, `
	  start <options>  | Start a Job that has a schedule.
	  stop  <options>  | Stop a Job
	  ls               | Status a Job -- currently only returns 'ACTIVE' jobs
	  get <options>    | Get a Job run status.

	  Call run <action> help for more on a sub-command
	`)
}

func (self *RunsTopLevel) Parse(args [] string) (exec CommandExec, err error) {

	defer func() {
		if r := recover(); r != nil {
			buf := new(bytes.Buffer)
			fmt.Fprintln(buf, r.(error).Error())
			fmt.Fprintf(buf, "\n %s usage:\n", self.subcommand)

			if self.task != nil {
				self.task.Usage(buf)
			}

			self.Usage(buf)
			err = errors.New(buf.String())
		}
	}()
	if len(args) == 0 {
		panic(errors.New("sub command required"))
	}
	logrus.Debugf("RunTopLevel args: %+v", args)
	self.subcommand = args[0]
	switch self.subcommand {
	case "ls":
		self.task = CommandParse(new(RunLs))
	case "get":
		self.task = CommandParse(new(RunStatusJob))
	case "start":
		self.task = CommandParse(new(RunStartJob))
	case "stop":
		self.task = CommandParse(new(RunStopJob))
	case "help", "--help":
		self.Usage(os.Stderr)
		return nil, errors.New("run usage")
	default:
		return nil, errors.New(fmt.Sprintf("run - don't understand '%s'\n", self.subcommand))
	}
	var subcommandArgs []string
	if len(args) > 1 {
		subcommandArgs = args[1:]
	}
	logrus.Debugf("run %s args: %+v", self.subcommand, subcommandArgs)
	if exec, err = self.task.Parse(subcommandArgs); err != nil {
		panic(err)
	} else {
		return exec, err
	}
}

// GET /v1/jobs/$jobId/runs
type RunLs JobId

func (self *RunLs) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("job ls", flag.ExitOnError)
	(*JobId)(self).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}

func (self *RunLs) Parse(args []string) (_ CommandExec, err error) {
	flags := flag.NewFlagSet("run ls", flag.ExitOnError)
	(*JobId)(self).FlagSet(flags)
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
	} else if err = (*JobId)(self).Validate(); err != nil {
		panic(err)
	} else {
		return self, nil
	}
}

func (self *RunLs) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.RunLs(string(*self))
}
// POST /v1/jobs/$jobId/runs
type RunStartJob JobId

func (self *RunStartJob) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("run start", flag.ExitOnError)
	(*JobId)(self).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}

func (self *RunStartJob) Parse(args []string) (_ CommandExec, err error) {
	flags := flag.NewFlagSet("run start", flag.ExitOnError)
	(*JobId)(self).FlagSet(flags)
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
	} else if err = (*JobId)(self).Validate(); err != nil {
		panic(err)
	} else {
		return self, nil
	}
}
func (self *RunStartJob) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.RunStartJob(string(*self))
}
// GET  /v1/jobs/$jobId/runs/$runId
type RunStatusJob struct {
	JobId
	RunId
}

func (self *RunStatusJob) FlagSet(flags *flag.FlagSet) *flag.FlagSet {
	self.JobId.FlagSet(flags)
	self.RunId.FlagSet(flags)
	return flags
}
func (self *RunStatusJob) Validate() error {
	if err := self.JobId.Validate(); err != nil {
		return err
	} else if err = self.RunId.Validate(); err != nil {
		return err
	}
	return nil
}

func (self *RunStatusJob) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("run status", flag.ExitOnError)
	self.FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}

func (self *RunStatusJob) Parse(args []string) (_ CommandExec, err error) {
	flags := flag.NewFlagSet("run status", flag.ExitOnError)
	self.FlagSet(flags)

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
	} else if err = self.Validate(); err != nil {
		panic(err)
	} else {
		return self, nil
	}
}
func (self *RunStatusJob) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.RunStatusJob(string(self.JobId), string(self.RunId))
}
// POST /v1/jobs/$jobId/runs/$runId/action/stop
type RunStopJob RunStatusJob

func (self *RunStopJob) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("run stop", flag.ExitOnError)
	(*RunStatusJob)(self).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}

func (self *RunStopJob) Parse(args []string) (_ CommandExec, err error) {
	flags := flag.NewFlagSet("run stop", flag.ExitOnError)
	(*RunStatusJob)(self).FlagSet(flags)
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
	} else if err = (*RunStatusJob)(self).Validate(); err != nil {
		panic(err)
	}
	return self, nil
}
func (self *RunStopJob) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.RunStopJob(string(self.JobId), string(self.RunId))
}
