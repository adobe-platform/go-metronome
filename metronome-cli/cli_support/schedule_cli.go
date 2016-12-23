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

// Scheduling related cli support structures implementing CommandParse and CommandExecute to run Metronome scheduling related
// API calls.
// These structures and implementations also implement the Value interface used by flag

// JobSched - schedules need a job-id and a schedule
type JobSched struct {
	JobID
	met.Schedule
}
// JobSchedRun - almost same as JobSched but with different Executor
type JobSchedRun struct {
	*JobSched
}
//
// SchedTopLevel - top-level, cli menu for scheduling.  generally parses out action before instantiating specific, action relate CommandParse implementation
//
type SchedTopLevel struct {
	subcommand string
	task       CommandParse
}
// Usage - schedule toplevel usage
func (theSchedule *SchedTopLevel) Usage(writer io.Writer) {
	fmt.Fprintf(writer, "schedule {create|delete|update|get|ls}  \n")
	fmt.Fprintln(writer, `
	  create  <options>  | Create a Schedule for a Job
	  delete  <options>  | Delete a Schedule for a Job
	  update  <options>  | Update a Schedule for a Job
	  get     <options>  | Get a single Schedule for a Job
	  ls                 | Get all Schedules for a Job
	`)
}
// Parse - parses out actions, delegates deeper parsing to action specific CommandParse implementations
func (theSchedule *SchedTopLevel) Parse(args [] string) (exec CommandExec, err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := new(bytes.Buffer)
			fmt.Fprintln(buf, r.(error).Error())
			fmt.Fprintf(buf, "\nschedule %s usage:\n", theSchedule.subcommand)

			if theSchedule.task != nil {
				theSchedule.task.Usage(buf)
			}
			theSchedule.Usage(buf)
			err = errors.New(buf.String())
		}
	}()
	logrus.Debugf("ScheduleTopLevel args: %+v", args)

	if len(args) == 0 {
		panic(errors.New("sub command required"))
	}

	theSchedule.subcommand = args[0]
	switch theSchedule.subcommand {
	case "create":
		// POST /v1/jobs/$jobId/schedules
		theSchedule.task = CommandParse(new(JobScheduleCreate))
	case "ls":
		// GET /v1/jobs/$jobId/schedules
		theSchedule.task = CommandParse(new(JobScheduleList))

	case "delete":
		// DELETE /v1/jobs/$jobid/schedules/$scheduleId
		theSchedule.task = CommandParse(new(JobSchedDelete))
	case "get":
		// GET /v1/jobs/$jobId/schedules/$scheduleId
		theSchedule.task = CommandParse(new(JobSchedGet))
	case "update":
		// PUT /v1/jobs/$jobId/schedules/$scheduleId
		theSchedule.task = CommandParse(new(JobSchedUpdate))
	case "help", "--help":
		panic(errors.New("Please help"))
	default:
		panic(fmt.Errorf("schedule Don't understand '%s'", theSchedule.subcommand))
	}
	var subcommandArgs []string
	if len(args) > 1 {
		subcommandArgs = args[1:]
	}
	logrus.Debugf("schedule %s args: %+v  task: %+v", theSchedule.subcommand, subcommandArgs, theSchedule.task)
	exec, err = theSchedule.task.Parse(subcommandArgs)
	if err != nil {
		panic(fmt.Errorf("SchedTopLevel parse failed %+v", err))
	}
	logrus.Debugf("SchedTopLevel parse succeeded %+v", exec)
	return exec, nil

}
// JobSchedBase -  collects parsing behavior needed in child classes
type JobSchedBase struct {
	JobID
	SchedID
}
// FlagSet - general flags (job-id,schedule-id). Delegates flag creation to JobID, SchedID
func (theSched *JobSchedBase) FlagSet(flags *flag.FlagSet) *flag.FlagSet {
	theSched.JobID.FlagSet(flags)
	theSched.SchedID.FlagSet(flags)
	return flags
}
// Validate - ensures we have valid schedule-id, job-id before returning executor
func (theSched *JobSchedBase) Validate() error {
	if err := theSched.JobID.Validate(); err != nil {
		return err
	} else if err = theSched.SchedID.Validate(); err != nil {
		return err
	}
	return nil
}
// JobSchedGet - CommandParse/CommandExecutor for runnig metronome
//    GET /v1/jobs/$jobId/schedules/$scheduleId
//    thin type based on JobSchedBase
type JobSchedGet JobSchedBase

// Usage - gets flags from JobSchedBase
func (theSched *JobSchedGet) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("schedule get", flag.ExitOnError)
	(*JobSchedBase)(theSched).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}
// Parse - parses JobSchedBase params but returns self as executor.  Converts all panics into errors
func (theSched *JobSchedGet) Parse(args []string) (_ CommandExec, err error) {
	flags := flag.NewFlagSet("schedule get", flag.ExitOnError)
	(*JobSchedBase)(theSched).FlagSet(flags)
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
	} else if err = (*JobSchedBase)(theSched).Validate(); err != nil {
		panic(err)
	} else {
		return theSched, nil
	}
}
// Execute - Runs GET /v1/jobs/$jobId/schedules/$scheduleId
func (theSched *JobSchedGet) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.GetSchedule(string(theSched.JobID), string(theSched.SchedID))
}
// JobSchedDelete - cli structure for executing Metronome API `DELETE /v1/jobs/$jobId/schedules/$scheduleId`
type JobSchedDelete JobSchedBase

// Usage - based on on JobSchedBase flags
func (theSched *JobSchedDelete) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("schedule delete", flag.ExitOnError)
	(*JobSchedBase)(theSched).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}
// Parse - parses cli flags using JobSchedBase but returns self as CommandExecution interface implementor
func (theSched *JobSchedDelete) Parse(args []string) (_ CommandExec, err error) {
	flags := flag.NewFlagSet("schedule delete", flag.ExitOnError)
	(*JobSchedBase)(theSched).FlagSet(flags)
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
	} else if err = (*JobSchedBase)(theSched).Validate(); err != nil {
		panic(err)
	} else {
		return theSched, nil
	}
}
// Execute - runs GET /v1/jobs/$jobId/schedules/$scheduleId
func (theSched *JobSchedDelete) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.DeleteSchedule(string(theSched.JobID), string(theSched.SchedID))
}
// JobScheduleList - cli structure to list a job's schedules -> GET /v1/jobs/$jobId/schedules
type JobScheduleList JobID

// Usage - flags come for job id
func (theSched *JobScheduleList) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("schedule ls", flag.ExitOnError)
	(*JobID)(theSched).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()

}
// Parse - flags parsed as with JobID but returns self as CommandExecutor on success
func (theSched *JobScheduleList) Parse(args [] string) (_ CommandExec, err error) {
	flags := flag.NewFlagSet("schedule ls", flag.ExitOnError)
	(*JobID)(theSched).FlagSet(flags)
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
	} else if err = (*JobID)(theSched).Validate(); err != nil {
		panic(err)
	} else {
		return theSched, nil
	}
	return theSched, nil
}
// Execute - implement CommandExec
//  - Runs GET /v1/jobs/$jobId/schedules
func (theSched *JobScheduleList) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.Schedules(string(*theSched))
}

// JobScheduleCreate - cli implementation of CommandParse,CommandExec to run -> POST /v1/jobs/$jobId/schedules
//   - derives from JobSched
type JobScheduleCreate JobSched

// Usage - uses flagset for JobSched
func (theSched *JobScheduleCreate) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("schedule create", flag.ExitOnError)
	(*JobSched)(theSched).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}
// Parse - parses flags per JobSched (job-id) but returns self as CommandExec with validation
func (theSched *JobScheduleCreate) Parse(args [] string) (_ CommandExec, err error) {
	logrus.Debugf("JobScheduleCreate.parse args: %+v", args)
	flags := flag.NewFlagSet("schedule create", flag.ExitOnError)
	(*JobSched)(theSched).FlagSet(flags)

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
	} else if err = (*JobSched)(theSched).Validate(); err != nil {
		panic(err)
	} else {
		return theSched, nil
	}
}
// Execute  - implement CommandExec.  Executes POST /v1/jobs/$jobId/schedules
func (theSched *JobScheduleCreate) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.CreateSchedule(string(theSched.JobID), &theSched.Schedule)
}

// JobSchedUpdate - cli type support executing PUT /v1/jobs/$jobId/schedules/$scheduleId
type JobSchedUpdate JobSched

// Usage - JobSched usage i.e. job-id, sched-id required
func (theSched *JobSchedUpdate) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("schedule update", flag.ExitOnError)
	(*JobSched)(theSched).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}

// Parse - JobSched flags but returns self as CommandExec when valid
func (theSched *JobSchedUpdate) Parse(args []string) (_ CommandExec, err error) {
	logrus.Debugf("JobSchedUpdate.Parse args: %s", args)
	flags := flag.NewFlagSet("schedule update", flag.ExitOnError)
	(*JobSched)(theSched).FlagSet(flags)
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
	} else if err = (*JobSched)(theSched).Validate(); err != nil {
		panic(err)
	} else {
		return theSched, nil
	}
}
// Execute - executes PUT /v1/jobs/$jobId/schedules/$scheduleId
func (theSched *JobSchedUpdate) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.UpdateSchedule(string(theSched.JobID), string(theSched.Schedule.ID), &theSched.Schedule)
}

// FlagSet - The JobSched flags needed to create a new Metronome schedule
func (theSched *JobSched) FlagSet(flags *flag.FlagSet) *flag.FlagSet {
	theSched.JobID.FlagSet(flags)
	flags.StringVar(&theSched.Schedule.ID, "sched-id", "", "Schedule Id")
	flags.StringVar(&theSched.Schedule.Cron, "cron", "", "Schedule Cron")
	flags.StringVar(&theSched.Schedule.Timezone, "tz", "GMT", "Schedule time zone")
	flags.IntVar(&theSched.Schedule.StartingDeadlineSeconds, "start-deadline", 0, "Schedule deadline")
	flags.StringVar(&theSched.Schedule.ConcurrencyPolicy, "concurrency-policy", "ALLOW", "Schedule concurrency.  One of ALLOW,FORBID,REPLACE")
	flags.BoolVar(&theSched.Schedule.Enabled, "enabled", true, "Enable the schedule")
	return flags
}
// Validate - validates that schedule and target
func (theSched *JobSched) Validate() error {
	if theSched.JobID == "" {
		return errors.New("Missing JobId in JobScheduleCreate")
	} else if theSched.Schedule.ID == "" {
		return errors.New("Missing SchedId in JobScheduleCreate")
	} else if theSched.Schedule.Cron == "" {
		return errors.New("Missing Cron in JobScheduleCreate")
	} else if !In(theSched.Schedule.ConcurrencyPolicy, []string{"ALLOW", "FORBID", "REPLACE"}) {
		return errors.New("Missing concurrency policy")
	} else if theSched.Schedule.StartingDeadlineSeconds < 2 {
		return errors.New("-starting-deadline-seconds must be > 1")
	}

	return nil
}

