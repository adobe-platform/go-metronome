package cli

import (
	met "github.com/adobe-platform/go-metronome/metronome"
	"github.com/Sirupsen/logrus"
	"fmt"
	"errors"
	"bytes"
	"io"
	"os"
	"flag"
)

//
// jobs top level cli parse/execute
//

// JobTopLevel - Top level type used to provide the `job <action>` functionality.
//   Implements CommandParse interface
type JobTopLevel struct {
	subcommand string
	task       CommandParse
}

// Usage - show usage
func (theJob *JobTopLevel) Usage(writer io.Writer) {
	fmt.Fprintf(writer, "job {create|delete|update|ls|get|schedules|schedule|help}\n")
	fmt.Fprintln(writer, `
	  create  <options>   | creates a Job
	  delete  <options>   | deletes a Job
	  update  <options>   | update a Job
	  get     <options>   | get a Job by job-id
	  schedules <options> | get all schedules [] for a Job
	  schedule  <options> | get a particular Schedule for Job
	  ls                  | get all Jobs []
	  Call job <action> help for more on a sub-command
	`)

}
// Parse - parse top-level commands.
func (theJob *JobTopLevel) Parse(args [] string) (exec CommandExec, err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := new(bytes.Buffer)
			fmt.Fprintln(buf, r.(error).Error())
			fmt.Fprintf(buf, "\njob %s usage:\n", theJob.subcommand)
			if theJob.task != nil {
				theJob.task.Usage(buf)
			}
			theJob.Usage(buf)
			err = errors.New(buf.String())
		}
	}()

	if len(args) == 0 {
		panic(errors.New("job subcommand required"))
	}
	logrus.Debugf("JobTopLevel job args: %+v\n", args)
	theJob.subcommand = args[0]
	switch  theJob.subcommand{
	case "create":
		// POST /v1/jobs
		x := CommandParse(new(JobCreateRuntime))
		theJob.task = x

	case "delete":
		// DELETE /v1/jobs/$jobid
		theJob.task = CommandParse(new(JobDelete))
	case "ls":
		// GET /v1/jobs
		theJob.task = CommandParse(new(JobList))
	case "get":
		// GET /v1/jobs/$jobId
		theJob.task = CommandParse(new(JobGet))
	case "update":
		// PUT /v1/jobs/$jobId
		theJob.task = CommandParse(new(JobUpdate))
	case "schedules":
		// GET /v1/jobs/$jobId/schedules  []Schedule
		theJob.task = CommandParse(new(JobScheduleList))
	case "schedule":
		theJob.task = CommandParse(new(JobScheduleCreate))
	case "help", "--help":
		theJob.Usage(os.Stderr)
		return nil, errors.New("job usage")
	default:
		return nil, fmt.Errorf("don't understand option %s", theJob.subcommand)
	}
	var subcommandArgs []string
	if len(args) > 1 {
		subcommandArgs = args[1:]
	}
	logrus.Debugf("job %s args: %+v\n", theJob.subcommand, subcommandArgs)

	if exec, err = theJob.task.Parse(subcommandArgs); err != nil {
		panic(err)
	} else {
		return exec, err
	}

}

// JobCreateConfig - provides backing structure needed to create a job via the command line.
//  Maps to many flags
//  Used in several major command functions: `job create` and `job update`
//    - Factored to support both
//  Used with Metronome **POST /v1/jobs**
type JobCreateConfig struct {
	JobID

	cpus                  float64
	disk                  int
	mem                   int
	description           string
	dockerImage           string
	restartPolicy         string
	activeDeadlineSeconds int
	constraints           ConstraintList
	volumes               VolumeList
	env                   NvList
	labels                LabelList
	artifacts             ArtifactList
	args                  RunArgs
	cmd                   string
	user                  string
	maxLaunchDelay        int
	runNow                bool
}
// makeJob - construct a metronome job for the structure - usually populated via cli flags
func (theJob *JobCreateConfig) makeJob() (*met.Job, error) {
	var container *met.Docker
	if theJob.dockerImage != "" {
		container = &met.Docker{
			Image: theJob.dockerImage,
		}
	}
	run, err := met.NewRun(theJob.cpus, theJob.disk, theJob.mem)

	if err != nil {
		return nil, err
	}
	if theJob.maxLaunchDelay < 1 {
		return nil, errors.New("max-launch-delay must be greater than 1")
	}
	run.SetMaxLaunchDelay(theJob.maxLaunchDelay)

	if len(theJob.constraints) > 0 {
		run.SetPlacement(&met.Placement{Constraints: []met.Constraint(theJob.constraints)})
	}
	if len(theJob.env) > 0 {
		run.SetEnv(theJob.env)
	}
	if len(theJob.args) > 0 {
		run.SetArgs([]string(theJob.args))
	}
	if len(theJob.volumes) > 0 {
		run.SetVolumes([]met.Volume(theJob.volumes))
	}
	if len(theJob.artifacts) > 0 {
		run.SetArtifacts([]met.Artifact(theJob.artifacts))
	}

	var description string
	if theJob.description != "" {
		description = theJob.description
	}
	var ll *met.Labels
	if theJob.labels.Location != "" || theJob.labels.Owner != "" {
		ll = (*met.Labels)(&theJob.labels)
	}
	if len(theJob.restartPolicy) > 0 || theJob.activeDeadlineSeconds != 0 {
		restart, err := met.NewRestart(theJob.activeDeadlineSeconds, theJob.restartPolicy)
		if err != nil {
			return nil, err
		}
		run.SetRestart(restart)

	}
	newJob, err := met.NewJob(string(theJob.JobID), description, ll, run)
	if err != nil {
		return nil, err

	} else if container != nil {
		newJob.GetRun().SetDocker(container).SetCmd(theJob.cmd)
	}
	logrus.Debugf("JobCreateRuntime: %+v", theJob)
	return newJob, nil

}
// JobCreateRuntime - Entity used to create a metronome job
//  Implements the CommandParse interface
type JobCreateRuntime struct {
	JobCreateConfig
	job           *met.Job
	disableRunNow bool
}
// FlagSet - Flags to setup creating a job.
//  - Populates JobCreateConfig with values
//
func (theJob *JobCreateRuntime) FlagSet(flags *flag.FlagSet) *flag.FlagSet {
	if theJob.env == nil {
		theJob.env = make(map[string]string)
	}

	logrus.Debugf("nvlist: %+v", theJob.env)
	flags.StringVar((*string)(&theJob.JobID), "job-id", "", "Job Id")
	flags.StringVar(&theJob.description, "description", "", "Job Description - optional")
	flags.StringVar((*string)(&theJob.dockerImage), "docker-image", "", "Docker Image")
	flags.Float64Var(&theJob.cpus, "cpus", DefaultCPUs, "cpus")
	flags.IntVar(&theJob.mem, "memory", DefaultMemory, "memory")
	flags.IntVar(&theJob.disk, "disk", DefaultDisk, "disk")
	flags.StringVar(&theJob.restartPolicy, "restart-policy", "NEVER", "Restart policy on job failure: NEVER or ALWAYS")
	flags.IntVar(&theJob.activeDeadlineSeconds, "restart-active-deadline-seconds", 0, "If the job fails, how long should we try to restart the job. If no value is set, this means forever.")
	flags.Var(&theJob.constraints, "constraint", "Add Constraint used to construct Job->Run->[]Constraint")
	flags.Var(&theJob.volumes, "volume", "/host:/container:{RO|RW} . Adds Volume passed to metrononome->Job->Run->Volumes. You can call more than once")
	flags.Var(&theJob.artifacts, "artifact", `uri=xxx  executable={true|false}  cache={true|false} extract={true|false} executable={true|false}
	                                cache,extract,executable are optional.  uri is required`)
	flags.Var(&theJob.args, "arg", "Adds Arg metrononome->Job->Run->Args. You can call more than once")
	flags.Var(&theJob.env, "env", "VAR=VAL . Adds Volume passed on to Job.Run.[]Volumes.  You can call more than once")
	flags.Var(&theJob.labels, "label", "Location=xxx; Owner=yyy")
	flags.StringVar(&theJob.user, "user", "root", "user to run as")
	flags.StringVar(&theJob.cmd, "cmd", "", "Command to run")
	flags.IntVar(&theJob.maxLaunchDelay, "max-launch-delay", 900, "Max Launch delay.  minimum 1")
	if !theJob.disableRunNow {
		flags.BoolVar(&theJob.runNow, "run-now", false, "Run this job now, otherwise it is created as unscheduled")
	}
	return flags
}

// Validate - validate the the structure can be used to create a job
func (theJob *JobCreateRuntime) Validate() error {
	if theJob.JobID == "" {
		return errors.New("Missing JobId")
	} else if theJob.cmd == "" && theJob.dockerImage == "" {
		return errors.New("Need command or docker image")
	} else if theJob.cpus <= 0.0 || theJob.mem <= 0 || theJob.disk <= 0 {
		return errors.New("cpus, memory, and disk must all be > 0")
	}
	return nil
}
// Usage - dump flag usage
func (theJob *JobCreateRuntime) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("job create", flag.ExitOnError)
	theJob.FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}

// JobRunNow - lightweight type to implement similar functionality to job create with the the target of running it immediatelu
//  - Implement CommandExec
type JobRunNow struct {
	job *met.Job
}
// Parse -
func (theJob *JobCreateRuntime) Parse(args []string) (exec CommandExec, err error) {
	logrus.Debugf("JobCreateRuntime.Parse %+v", args)
	flags := flag.NewFlagSet("job create", flag.ExitOnError)
	theJob.FlagSet(flags)

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
	} else if err = theJob.Validate(); err != nil {
		panic(err)
	}
	theJob.job, err = theJob.JobCreateConfig.makeJob()
	if err != nil {
		return nil, err
	}
	if theJob.runNow {
		return &JobRunNow{job:theJob.job}, nil
	}
	return theJob, nil

}
// Execute - create and execute a job
func (theJob *JobRunNow) Execute(runtime *Runtime) (interface{}, error) {
	logrus.Debugf("JobCreateRuntime.Execute %+v", runtime)
	_, err := runtime.client.CreateJob(theJob.job)
	if err != nil {
		return nil, err

	}
	return runtime.client.StartJob(theJob.job.ID)
}

// Execute - create a job
func (theJob *JobCreateRuntime) Execute(runtime *Runtime) (interface{}, error) {
	logrus.Debugf("JobCreateRuntime.Execute %+v", runtime)
	return runtime.client.CreateJob(theJob.job)
}

// JobDelete - Implement CommandParse and CommandExecute
// DELETE /v1/jobs/$jobId
type JobDelete JobID

// Usage - CommandParse implementation/
func (theJob *JobDelete) Usage(writer io.Writer) {
	fmt.Fprintf(writer, "job delete:\n")
	flags := flag.NewFlagSet("job delete", flag.ExitOnError)
	(*JobID)(theJob).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}

// Parse - Handle command line flags.
//  - CommandParse implementation
func (theJob *JobDelete) Parse(args []string) (exec CommandExec, err error) {
	flags := flag.NewFlagSet("job delete", flag.ExitOnError)
	(*JobID)(theJob).FlagSet(flags)

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
	} else if err = (*JobID)(theJob).Validate(); err != nil {
		panic(err)
	} else {
		return theJob, nil
	}
}
// Execute - delete the job
func (theJob *JobDelete) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.DeleteJob((string)(*theJob))
}

// JobGet - Get a job via command line.
//   - Implements CommandParse & CommandExecute interfaces
//   - GET /v1/jobs/$jobId
type JobGet JobID

// Usage - CommandParse implementation
func (theJob *JobGet) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("job get", flag.ExitOnError)
	(*JobID)(theJob).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}
// Parse - the command line flags
func (theJob *JobGet) Parse(args []string) (exec CommandExec, err error) {
	flags := flag.NewFlagSet("job get", flag.ExitOnError)
	(*JobID)(theJob).FlagSet(flags)
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
	} else if err = (*JobID)(theJob).Validate(); err != nil {
		panic(err)
	} else {
		return theJob, nil
	}
}
// Execute - get the job from metronome
func (theJob *JobGet) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.GetJob(string(*theJob))
}

// JobList - no arg type to list all the jobs in the system via command line
//  - Implements CommandParse/CommandExecute interfaces
//  - GET /v1/jobs
type JobList int

// Usage - CommandParse implementation
func (theJob *JobList) Usage(writer io.Writer) {
	fmt.Fprintf(writer, "job ls\n\tList all jobs\n")
}
// Parse - nothing to parse.  Implements CommandParse
func (theJob *JobList) Parse([] string) (CommandExec, error) {
	return theJob, nil
}
// Execute - get the jobs from Metronome
func (theJob *JobList) Execute(runtime *Runtime) (interface{}, error) {
	jobs, err := runtime.client.Jobs()
	if err != nil {
		return nil, err
	}
	return jobs, nil

}
// JobUpdate - update a job via the command line
//  - Implements CommandParse/CommandExecute
// -  PUT /v1/jobs/$jobId
type JobUpdate JobCreateRuntime

// Usage - CommandParse implementation
func (theJob *JobUpdate) Usage(writer io.Writer) {
	theJob.disableRunNow = true
	flags := flag.NewFlagSet("job update", flag.ExitOnError)
	// must cast to JobRuntime or go chooses JobId.Flagset...
	(*JobCreateRuntime)(theJob).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}

// Parse - implement CommandParse.
//   Use underlying (actual) JobCreateRuntime
func (theJob *JobUpdate) Parse(args [] string) (_ CommandExec, err error) {
	theJob.disableRunNow = true
	flags := flag.NewFlagSet("job update", flag.ExitOnError)
	(*JobCreateRuntime)(theJob).FlagSet(flags)
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
	} else if err = (*JobCreateRuntime)(theJob).Validate(); err != nil {
		panic(err)
	} else {
		theJob.job, err = theJob.JobCreateConfig.makeJob()
		if err != nil {
			return nil, err
		}
		return theJob, nil

	}
}
// Execute - implement CommandExec
func (theJob *JobUpdate) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.UpdateJob(string(theJob.JobID), theJob.job)
}
