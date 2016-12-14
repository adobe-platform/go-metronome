package cli

import "flag"
import (
	met "github.com/adobe-platform/go-metronome/metronome"
	"github.com/Sirupsen/logrus"
	"fmt"
	"errors"
	"bytes"
	"io"
)


// JobSched
type JobSched struct {
	JobId
	met.Schedule
}
type JobSchedRun struct {
	*JobSched
}
//
// Schedule
//
type SchedTopLevel struct {
	subcommand string
	task       CommandParse
}

func (self *SchedTopLevel) Usage(writer io.Writer) {
	fmt.Fprintf(writer, "schedule {create|delete|update|get|ls}  \n")
	fmt.Fprintln(writer, `
	  create  <options>  | Create a Schedule for a Job
	  delete  <options>  | Delete a Schedule for a Job
	  update  <options>  | Update a Schedule for a Job
	  get     <options>  | Get a single Schedule for a Job
	  ls                 | Get all Schedules for a Job
	`)
}
func (self *SchedTopLevel) Parse(args [] string) (exec CommandExec, err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := new(bytes.Buffer)
			fmt.Fprintln(buf, r.(error).Error())
			fmt.Fprintf(buf, "\nschedule %s usage:\n", self.subcommand)

			if self.task != nil {
				self.task.Usage(buf)
			}
			self.Usage(buf)
			err = errors.New(buf.String())
		}
	}()
	logrus.Debugf("ScheduleTopLevel args: %+v", args)

	if len(args) == 0 {
		panic(errors.New("sub command required"))
	}

	self.subcommand = args[0]
	switch self.subcommand {
	case "create":
		// POST /v1/jobs/$jobId/schedules
		self.task = CommandParse(new(JobScheduleCreate))
	case "ls":
		// GET /v1/jobs/$jobId/schedules
		self.task = CommandParse(new(JobScheduleList))

	case "delete":
		// DELETE /v1/jobs/$jobid/schedules/$scheduleId
		self.task = CommandParse(new(JobSchedDelete))
	case "get":
		// GET /v1/jobs/$jobId/schedules/$scheduleId
		self.task = CommandParse(new(JobSchedGet))
	case "update":
		// PUT /v1/jobs/$jobId/schedules/$scheduleId
		self.task = CommandParse(new(JobSchedUpdate))
	case "help", "--help":
		panic(errors.New("Please help"))
	default:
		panic(errors.New(fmt.Sprintf("schedule Don't understand '%s'\n", self.subcommand)))
	}
	var subcommandArgs []string
	if len(args) > 1 {
		subcommandArgs = args[1:]
	}
	logrus.Debugf("schedule %s args: %+v  task: %+v", self.subcommand, subcommandArgs, self.task)
	if exec, err = self.task.Parse(subcommandArgs); err != nil {
		fmt.Errorf("SchedTopLevel parse failed %+v\n", err)
		panic(err)
	} else {
		fmt.Errorf("SchedTopLevel parse succeeded %+v\n", exec)
		return exec, nil
	}
}
// JobSchedBase collects parsing behavior needed in child classes
type JobSchedBase struct {
	JobId
	SchedId
}

func (self *JobSchedBase) FlagSet(flags *flag.FlagSet) *flag.FlagSet {
	self.JobId.FlagSet(flags)
	self.SchedId.FlagSet(flags)
	return flags
}
func (self *JobSchedBase) Validate() error {
	if err := self.JobId.Validate(); err != nil {
		return err
	} else if err = self.SchedId.Validate(); err != nil {
		return err
	}
	return nil
}
// GET /v1/jobs/$jobId/schedules/$scheduleId
type JobSchedGet JobSchedBase

func (self *JobSchedGet) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("schedule get", flag.ExitOnError)
	(*JobSchedBase)(self).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}

func (self *JobSchedGet) Parse(args []string) (_ CommandExec, err error) {
	flags := flag.NewFlagSet("schedule get", flag.ExitOnError)
	(*JobSchedBase)(self).FlagSet(flags)
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
	} else if err = (*JobSchedBase)(self).Validate(); err != nil {
		panic(err)
	} else {
		return self, nil
	}
}
func (self *JobSchedGet) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.GetSchedule(string(self.JobId), string(self.SchedId))
}
// DELETE /v1/jobs/$jobId/schedules/$scheduleId
type JobSchedDelete JobSchedBase

func (self *JobSchedDelete) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("schedule delete", flag.ExitOnError)
	(*JobSchedBase)(self).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}
func (self *JobSchedDelete) Parse(args []string) (_ CommandExec, err error) {
	flags := flag.NewFlagSet("schedule delete", flag.ExitOnError)
	(*JobSchedBase)(self).FlagSet(flags)
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
	} else if err = (*JobSchedBase)(self).Validate(); err != nil {
		panic(err)
	} else {
		return self, nil
	}
}

func (self *JobSchedDelete) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.DeleteSchedule(string(self.JobId), string(self.SchedId))
}
// GET /v1/jobs/$jobId/schedules
type JobScheduleList JobId

func (self *JobScheduleList) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("schedule ls", flag.ExitOnError)
	(*JobId)(self).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()

}
func (self *JobScheduleList) Parse(args [] string) (_ CommandExec, err error) {
	flags := flag.NewFlagSet("schedule ls", flag.ExitOnError)
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
	return self, nil
}
// JobScheduleList - implement CommandExec
func (self *JobScheduleList) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.Schedules(string(*self))
}

// POST /v1/jobs/$jobId/schedules
type JobScheduleCreate JobSched

func (self *JobScheduleCreate) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("schedule create", flag.ExitOnError)
	(*JobSched)(self).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}
func (self *JobScheduleCreate) Parse(args [] string) (_ CommandExec, err error) {
	logrus.Debugf("JobScheduleCreate.parse args: %+v", args)
	flags := flag.NewFlagSet("schedule create", flag.ExitOnError)
	(*JobSched)(self).FlagSet(flags)

	defer func() {
		if r := recover(); r != nil {
			buf := new(bytes.Buffer)
			flags.SetOutput(buf)
			fmt.Fprintln(buf, err.Error())
			err = errors.New(buf.String())
			panic(err)
		}
	}()
	if err = flags.Parse(args); err != nil {
		logrus.Debugf("JobScheduleCreate.parse failed %+v", err)
		panic(err)
	} else if err = (*JobSched)(self).Validate(); err != nil {
		panic(err)
	} else {
		return self, nil
	}
}
// JobScheduleCreate- implement CommandExec
func (self *JobScheduleCreate) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.CreateSchedule(string(self.JobId), &self.Schedule)
}

// PUT /v1/jobs/$jobId/schedules/$scheduleId
type JobSchedUpdate JobSched

func (self *JobSchedUpdate) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("schedule update", flag.ExitOnError)
	(*JobSched)(self).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}

func (self *JobSchedUpdate) Parse(args []string) (_ CommandExec, err error) {
	logrus.Debugf("JobSchedUpdate.Parse args: %s", args)
	flags := flag.NewFlagSet("schedule update", flag.ExitOnError)
	(*JobSched)(self).FlagSet(flags)
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
	} else if err = (*JobSched)(self).Validate(); err != nil {
		panic(err)
	} else {
		return self, nil
	}
}
func (self *JobSchedUpdate) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.UpdateSchedule(string(self.JobId), string(self.Schedule.ID), &self.Schedule)
}


func (self *JobSched) FlagSet(flags *flag.FlagSet) *flag.FlagSet {
	self.JobId.FlagSet(flags)
	flags.StringVar(&self.Schedule.ID, "sched-id", "", "Schedule Id")
	flags.StringVar(&self.Schedule.Cron, "cron", "", "Schedule Cron")
	flags.StringVar(&self.Schedule.Timezone, "tz", "GMT", "Schedule time zone")
	flags.IntVar(&self.Schedule.StartingDeadlineSeconds, "start-deadline", 0, "Schedule deadline")
	flags.StringVar(&self.Schedule.ConcurrencyPolicy, "concurrency-policy", "ALLOW", "Schedule concurrency.  One of ALLOW,FORBID,REPLACE")
	flags.BoolVar(&self.Schedule.Enabled, "enabled",true,"Enable the schedule")
	return flags
}
func (self *JobSched) Validate() error {
	if self.JobId == "" {
		return errors.New("Missing JobId in JobScheduleCreate")
	} else if self.Schedule.ID == "" {
		return errors.New("Missing SchedId in JobScheduleCreate")
	} else if self.Schedule.Cron == "" {
		return errors.New("Missing Cron in JobScheduleCreate")
	} else if !In(self.Schedule.ConcurrencyPolicy, []string{"ALLOW", "FORBID", "REPLACE"}) {
		return errors.New("Missing concurrency policy")
	} else if self.Schedule.StartingDeadlineSeconds < 2 {
		return errors.New("-starting-deadline-seconds must be > 1")
	}

	return nil
}

