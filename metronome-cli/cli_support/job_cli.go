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
type JobTopLevel struct {
	subcommand string
	task       CommandParse
}

func (self *JobTopLevel) Usage(writer io.Writer) {
	fmt.Fprintf(writer, "job {create|delete|update|ls|get|schedules|schedule|help}\n")
}
func (self *JobTopLevel) Parse(args [] string) (exec CommandExec, err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := new(bytes.Buffer)
			fmt.Fprintln(buf, r.(error).Error())
			fmt.Fprintf(buf, "\njob %s usage:\n", self.subcommand)
			if self.task != nil {
				self.task.Usage(buf)
			}
			self.Usage(buf)
			err = errors.New(buf.String())
		}
	}()

	if len(args) == 0 {
		panic(errors.New("job subcommand required"))
	}
	logrus.Debugf("JobTopLevel job args: %+v\n", args)
	self.subcommand = args[0]
	switch  self.subcommand{
	case "create":
		// POST /v1/jobs
		x := CommandParse(new(JobCreateRuntime))
		self.task = x

	case "delete":
		// DELETE /v1/jobs/$jobid
		self.task = CommandParse(new(JobDelete))
	case "ls":
		// GET /v1/jobs
		self.task = CommandParse(new(JobList))
	case "get":
		// GET /v1/jobs/$jobId
		self.task = CommandParse(new(JobGet))
	case "update":
		// PUT /v1/jobs/$jobId
		self.task = CommandParse(new(JobUpdate))
	case "schedules":
		// GET /v1/jobs/$jobId/schedules  []Schedule
		self.task = CommandParse(new(JobScheduleList))
	case "schedule":
		self.task = CommandParse(new(JobScheduleCreate))
	case "help", "--help":
		self.Usage(os.Stderr)
		return nil, errors.New("job usage")
	default:
		return nil, errors.New(fmt.Sprintf("job Don't understand option '%s'\n", self.subcommand))
	}
	var subcommandArgs []string
	if len(args) > 1 {
		subcommandArgs = args[1:]
	}
	logrus.Debugf("job %s args: %+v\n", self.subcommand, subcommandArgs)

	if exec, err = self.task.Parse(subcommandArgs); err != nil {
		panic(err)
	} else {
		return exec, err
	}

}


// POST /v1/jobs
type JobCreateConfig struct {
	JobId

	cpus                    float64
	disk                    int
	mem                     int
	description             string
	docker_image            string
	restart_policy          string
	active_deadline_seconds int
	constraints             ConstraintList
	volumes                 VolumeList
	env                     NvList
	labels                  LabelList
	artifacts               ArtifactList
	args                    RunArgs
	cmd                     string
	user                    string
	maxLaunchDelay          int
	runNow                  bool
}

func (self *JobCreateConfig) makeJob() (*met.Job, error) {
	container := met.Docker{
		Image_: self.docker_image,
	}
	run, err := met.NewRun(self.cpus, self.disk, self.mem)

	if err != nil {
		return nil, err
	}
	if self.maxLaunchDelay < 1 {
		return nil, errors.New("max-launch-delay must be greater than 1")
	} else {
		run.SetMaxLaunchDelay(self.maxLaunchDelay)
	}
	if len(self.constraints) > 0 {
		run.SetPlacement(&met.Placement{Constraints_: []met.Constraint(self.constraints)})
	}
	if len(self.env) > 0 {
		run.SetEnv(self.env)
	}
	if len(self.args) > 0 {
		run.SetArgs([]string(self.args))
	}
	if len(self.volumes) > 0 {
		run.SetVolumes([]met.Volume(self.volumes))
	}
	if len(self.artifacts) > 0 {
		run.SetArtifacts([]met.Artifact(self.artifacts))
	}

	var description string
	if self.description != "" {
		description = self.description
	}
	var ll *met.Labels
	if self.labels.Location != "" || self.labels.Owner != "" {
		ll = (*met.Labels)(&self.labels)
	}
	if len(self.restart_policy) > 0 || self.active_deadline_seconds != 0 {
		if restart, err := met.NewRestart(self.active_deadline_seconds, self.restart_policy); err != nil {
			return nil, err
		} else {
			run.SetRestart(restart)
		}
	}
	newJob, err := met.NewJob(string(self.JobId), description, ll, run)
	if err != nil {
		return nil, err

	} else {
		newJob.Run().SetDocker(&container).SetCmd(self.cmd)
	}
	logrus.Debugf("JobCreateRuntime: %+v\n", self)
	return newJob, nil

}

type JobCreateRuntime struct {
	JobCreateConfig
	job *met.Job
}

func (self *JobCreateRuntime) FlagSet(flags *flag.FlagSet) *flag.FlagSet {
	if self.env == nil {
		self.env = make(map[string]string)
	}

	logrus.Debugf("nvlist: %+v\n", self.env)
	flags.StringVar((*string)(&self.JobId), "job-id", "", "Job Id")
	flags.StringVar(&self.description, "description", "", "Job Description - optional")
	flags.StringVar((*string)(&self.docker_image), "docker-image", DefaultImage, "Docker Image")
	flags.Float64Var(&self.cpus, "cpus", DefaultCPUs, "cpus")
	flags.IntVar(&self.mem, "memory", DefaultMemory, "memory")
	flags.IntVar(&self.disk, "disk", DefaultDisk, "disk")
	flags.StringVar(&self.restart_policy, "restart-policy", "NEVER", "Restart policy on job failure: NEVER or ALWAYS")
	flags.IntVar(&self.active_deadline_seconds, "restart-active-deadline-seconds", 0, "If the job fails, how long should we try to restart the job. If no value is set, this means forever.")
	flags.Var(&self.constraints, "constraint", "Add Constraint used to construct Job->Run->[]Constraint")
	flags.Var(&self.volumes, "volume", "/host:/container:{RO|RW} . Adds Volume passed to metrononome->Job->Run->Volumes. You can call more than once")
	flags.Var(&self.args, "arg", "Adds Arg metrononome->Job->Run->Args. You can call more than once")
	flags.Var(&self.env, "env", "VAR=VAL . Adds Volume passed to metrononome->Job->Run->Volumes.  You can call more than once")
	flags.Var(&self.labels, "label", "Location=xxx; Owner=yyy")
	flags.StringVar(&self.user, "user", "root", "user to run as")
	flags.StringVar(&self.cmd, "cmd", "", "Command to run")
	flags.IntVar(&self.maxLaunchDelay, "max-launch-delay", 900, "Max Launch delay.  minimum 1")
	flags.BoolVar(&self.runNow, "run-now", false, "Run this job now, otherwise it is created as unscheduled")
	return flags
}

func (self *JobCreateRuntime) Validate() error {
	if self.JobId == "" {
		return errors.New("Missing JobId")
	} else if self.cmd == "" && self.docker_image == "" {
		return errors.New("Need command or docker image")
	} else if self.cpus <= 0.0 || self.mem <= 0 || self.disk <= 0 {
		return errors.New("cpus, memory, and disk must all be > 0")
	}
	return nil
}
func (self *JobCreateRuntime) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("job create", flag.ExitOnError)
	self.FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}

func (self *JobCreateRuntime) Parse(args []string) (exec CommandExec, err error) {
	logrus.Debugf("JobCreateRuntime.Parse %+v\n", args)
	flags := flag.NewFlagSet("job create", flag.ExitOnError)
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
	}
	if self.job, err = self.JobCreateConfig.makeJob(); err != nil {
		return nil, err
	} else {
		return self, nil
	}
}

func (self *JobCreateRuntime) Execute(runtime *Runtime) (interface{}, error) {
	logrus.Debugf("JobCreateRuntime.Execute %+v\n", runtime)
	return runtime.client.CreateJob(self.job)
}

// DELETE /v1/jobs/$jobId

type JobDelete JobId

func (self *JobDelete) Usage(writer io.Writer) {
	fmt.Fprintf(writer, "job delete:\n")
	flags := flag.NewFlagSet("job delete", flag.ExitOnError)
	(*JobId)(self).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}

func (self *JobDelete) Parse(args []string) (exec CommandExec, err error) {
	flags := flag.NewFlagSet("job delete", flag.ExitOnError)
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

func (self *JobDelete) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.DeleteJob((string)(*self))
}
// GET /v1/jobs/$jobId
type JobGet JobId

func (self *JobGet) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("job get", flag.ExitOnError)
	(*JobId)(self).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}

func (self *JobGet) Parse(args []string) (exec CommandExec, err error) {
	flags := flag.NewFlagSet("job get", flag.ExitOnError)
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
func (self *JobGet) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.GetJob(string(*self))
}
// GET /v1/jobs
type JobList int

func (self *JobList) Usage(writer io.Writer) {
	fmt.Fprintf(writer, "job ls\n\tList all jobs\n")
}

func (self *JobList) Parse([] string) (CommandExec, error) {
	return self, nil
}
func (self *JobList) Execute(runtime *Runtime) (interface{}, error) {
	if jobs, err := runtime.client.Jobs(); err != nil {
		return nil, err
	} else {
		return jobs, nil
	}
}
// PUT /v1/jobs/$jobId
type JobUpdate JobCreateRuntime

func (self *JobUpdate) Usage(writer io.Writer) {
	flags := flag.NewFlagSet("job update", flag.ExitOnError)
	(*JobCreateRuntime)(self).FlagSet(flags)
	flags.SetOutput(writer)
	flags.PrintDefaults()
}

// JobUpdate - implement CommandParse
func (self *JobUpdate) Parse(args [] string) (_ CommandExec, err error) {
	flags := flag.NewFlagSet("job update", flag.ExitOnError)
	(*JobCreateRuntime)(self).FlagSet(flags)
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
	} else if err = (*JobCreateRuntime)(self).Validate(); err != nil {
		panic(err)
	} else {
		if self.job, err = self.JobCreateConfig.makeJob(); err != nil {
			return nil, err
		} else {
			return self, nil
		}
	}
}
// JobUpdate - implement CommandExec
func (self *JobUpdate) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.JobUpdate(string(self.JobId), self.job)
}
