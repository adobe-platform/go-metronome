package main

import (
	"fmt"
	met "github.com/adobe-platform/go-metronome/metronome"
	"github.com/Sirupsen/logrus"
	"flag"
	"os"
	"strings"
	"errors"
	"encoding/json"
	"io"
	"bytes"
)
// Command line defaults
const (
	DefaultHTTPAddr = "http://localhost:9000"
	DefaultImage = "alpine:3.4"
	DefaultCPUs = 0.2
	DefaultMemory = 128
	DefaultDisk = 128
)
//Runtime represents the global options passed to all CommandExec.Execute methods.
//In particular, it keeps the metronome client and the other useful global options
type Runtime struct {
	httpAddr string
	flags    *flag.FlagSet
	debug    bool
	help     bool
	client   met.Metronome
}
// CommandExec is an interface returned by Parse when options are successfully parsed
// receiver Execute is passed the global options include the Metronome client interface
type CommandExec interface {
	Execute(runtime *Runtime) (interface{}, error)
}
// CommandParse
// implementor are passed arguments.
type CommandParse interface {
	Parse(args []string) (CommandExec, error)
	Usage(writer io.Writer)
}
type CommandMap map[string]CommandParse

var commands CommandMap
//
// Job{Create|Update} take many parameters that must be validated and stored in nested structures
// These are set via flag.Var.  When using flag.Var, flag expects the passed pointer to implement the flag.Value interface
// So as to not effect the behavior of the actual types, these critical types are effectively aliased below to provide
// the correct command line handling for flag and the real type.  By doing so, it preserves the real types behavior
// flag.Var calls flag.Value interface of the provided interface{}
// The following light-weight types implement Value while preserving the Set/String symantics of the `real` type it alias.
type RunArgs []string
type NvList map[string]string
type ConstraintList [] met.Constraint
type VolumeList [] met.Volume
type LabelList  met.Labels
type ArtifactList  []met.Artifact

// type override to support parsing.  []string alias for met.Run.Args
// It implements flag.Value via Set/String

func (i *RunArgs) String() string {
	return fmt.Sprintf("%s", *i)
}
// The second method is Set(value string) error
func (i *RunArgs) Set(value string) error {
	logrus.Debugf("Args.Set %s\n", value)
	*i = append(*i, value)
	return nil
}
// type override to support parsing.  LabelList alias' met.Labels
// It implements flag.Value via Set/String

func (i *LabelList) String() string {
	return fmt.Sprintf("%s", *i)
}
// The second method is Set(value string) error
func (lb *LabelList) Set(value string) error {
	logrus.Debugf("LabelList %s\n", value)
	v := strings.Split(value, ";")
	logrus.Debugf("LabelList %+v\n", v)
	//lb := LabelList{}
	for _, ii := range v {
		nv := strings.Split(ii, "=")
		switch strings.ToLower(nv[0]) {
		case "location":
			lb.Location = nv[1]
		case "owner":
			lb.Owner = nv[1]
		default:
			return errors.New("Unknown value" + nv[0])
		}
	}
	if lb.Location == ""  && lb.Owner == "" {
		return errors.New("Missing both location and owner")
	}
	return nil
}
// type override to support parsing.  env alias' map[string]string
// It implements flag.Value via Set/String

func (i *NvList) String() string {
	return fmt.Sprintf("%s", *i)
}
// The second method is Set(value string) error
func (self *NvList) Set(value string) error {
	logrus.Debugf("NvList %+v %s\n", self, value)
	nv := strings.Split(value, "=")
	if len(nv) != 2 {
		return errors.New("Environment vars should be NAME=VALUE")
	}
	logrus.Debugf("NvList %+v\n", nv)
	vv := (*self)
	vv[nv[0]] = nv[1]
	return nil
}
// ConstraintList
// type override to support parsing.  ConstraintList alias' []met.Constraint
// It implements flag.Value via Set/String

func (i *ConstraintList) String() string {
	return fmt.Sprintf("%s", *i)
}
// The second method is Set(value string) error
func (i *ConstraintList) Set(value string) error {
	if con, err := met.StrToConstraint(value); err != nil {
		return err
	} else {
		*i = ConstraintList(append(*i, *con))
		return nil
	}
}
// type override to support parsing.  VolumeList alias' []met.Volume
// It implements flag.Value via Set/String

func (i *VolumeList) String() string {
	return fmt.Sprintf("%s", *i)
}
// The second method is Set(value string) error
func (i *VolumeList) Set(value string) error {
	pieces := strings.Split(value, ":")
	if len(pieces) == 3 {

	} else if len(pieces) == 2 {
		pieces = append(pieces, "RO")
	}
	if vol, err := met.NewVolume(pieces[0], pieces[1], pieces[2]); err != nil {
		return err
	} else {
		*i = append(*i, *vol)
	}
	return nil
}
// type override to support parsing.  ArtifactList alias' []met.Artifact
// It implements flag.Value via Set/String
func (i *ArtifactList) String() string {
	return fmt.Sprintf("%s", *i)
}
// The second method is Set(value string) error
func (i *ArtifactList) Set(value string) error {
	return nil
}
//
// Global flags are kept in 'Runtime'.  main takes care of sending Parse the correct list of args
//
func (self *Runtime) FlagSet(name  string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	flags.StringVar(&self.httpAddr, "metronome-url", DefaultHTTPAddr, "Set the Metronome address")
	flags.BoolVar(&self.debug, "debug", false, "Turn on debug")
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
//
// Support parsing re-use
// JobId, RunId, and SchedId are used in many calls.  Those 'types' are never used directly.  Instead they are part of other structs
//
// base class used to parse args for many commands requiring `job-id` parsing
type JobId string

func (self *JobId) FlagSet(flags *flag.FlagSet) *flag.FlagSet {
	flags.StringVar((*string)(self), "job-id", "", "Job Id")
	return flags
}
func (self *JobId) Validate() error {
	logrus.Debugf("JobId.Validate\n")
	if string(*self) == "" {
		return errors.New("job-id required")
	}
	return nil
}

type SchedId string

func (self *SchedId) FlagSet(flags *flag.FlagSet) *flag.FlagSet {
	flags.StringVar((*string)(self), "sched-id", "", "Schedule Id")
	return flags
}
func (self *SchedId) Validate() error {
	if string(*self) == "" {
		return errors.New("sched-id required")
	}
	return nil
}
// RunId is used in several REST calls
type RunId string

func (self *RunId) FlagSet(flags *flag.FlagSet) *flag.FlagSet {
	flags.StringVar((*string)(self), "run-id", "", "Run Id")
	return flags
}
func (self *RunId) Validate() error {
	if string(*self) == "" {
		return errors.New("run-id required")
	}
	return nil
}
// JobSched
type JobSched struct {
	JobId
	met.Schedule
}
type JobSchedRun struct {
	*JobSched
}

func in(val string, targ []string) bool {
	for _, cur := range targ {
		if cur == val {
			return true
		}
	}
	return false
}

func (self *JobSched) FlagSet(flags *flag.FlagSet) *flag.FlagSet {
	//flags := self.JobId.FlagSet(name)
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
	logrus.Debugf("JobSched.Validate\n")
	if self.JobId == "" {
		return errors.New("Missing JobId in JobScheduleCreate")
	} else if self.Schedule.ID == "" {
		return errors.New("Missing SchedId in JobScheduleCreate")
	} else if self.Schedule.Cron == "" {
		return errors.New("Missing Cron in JobScheduleCreate")
	} else if !in(self.Schedule.ConcurrencyPolicy, []string{"ALLOW", "FORBID", "REPLACE"}) {
		return errors.New("Missing concurrency policy")
	} else if self.Schedule.StartingDeadlineSeconds < 2 {
		return errors.New("-starting-deadline-seconds must be > 1")
	}

	return nil
}
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

type RunsTopLevel JobTopLevel

func (self *RunsTopLevel) Usage(writer io.Writer) {
	fmt.Fprintf(writer, "run <action> [options]:\n")
	fmt.Fprint(writer, `
	  start [options]
	  stop  [options]
	  ls
	  get [options]

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
	logrus.Debugf("RunTopLevel args: %+v\n", args)
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
	logrus.Debugf("run %s args: %+v\n", self.subcommand, subcommandArgs)
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
//
// Schedule
//
type SchedTopLevel struct {
	subcommand string
	task       CommandParse
}

func (self *SchedTopLevel) Usage(writer io.Writer) {
	fmt.Fprintf(writer, "schedule {create|delete|update|get|ls}  \n")
	fmt.Fprint(writer, `
	  create  <options>
	  delete  <options>
	  update  <options>
	  get     <options>
	  ls
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
	logrus.Debugf("ScheduleTopLevel args: %+v\n", args)

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
	logrus.Debugf("schedule %s args: %+v  task: %+v\n", self.subcommand, subcommandArgs, self.task)
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
	return runtime.client.JobScheduleGet(string(self.JobId), string(self.SchedId))
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
	return runtime.client.JobScheduleDelete(string(self.JobId), string(self.SchedId))
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
	return runtime.client.JobScheduleList(string(*self))
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
	logrus.Debugf("JobScheduleCreate.parse args: %+v\n", args)
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
		logrus.Debugf("JobScheduleCreate.parse failed %+v\n", err)
		panic(err)
	} else if err = (*JobSched)(self).Validate(); err != nil {
		panic(err)
	} else {
		return self, nil
	}
}
// JobScheduleCreate- implement CommandExec
func (self *JobScheduleCreate) Execute(runtime *Runtime) (interface{}, error) {
	logrus.Debugf("JobScheduleCreate.Execute\n")
	return runtime.client.JobScheduleCreate(string(self.JobId), &self.Schedule)
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
	logrus.Debugf("JobSchedUpdate.Parse args;: %s\n", args)
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
	return runtime.client.JobScheduleUpdate(string(self.JobId), string(self.Schedule.ID), &self.Schedule)
}
// initialize the top level command map
func init() {
	commands = CommandMap{
		"job": CommandParse(new(JobTopLevel)),
		"run": CommandParse(new(RunsTopLevel)),
		"schedule": CommandParse(new(SchedTopLevel)),
		"metrics": CommandParse(new(Metrics)),
		"ping": CommandParse(new(Ping)),
	}
}

func Usage(msg string) {
	if msg != "" {
		logrus.Errorf(" %s ", msg)
	}
	logrus.Errorf("usage: %s <global-options> <action: one of {%s}> [<action options>|help ]", os.Args[0], strings.Join([]string{
		"job",
		"run",
		"schedule",
		"metrics",
		"ping",
		"help",
	}, "|"))
	fmt.Println(" For more help, use ")
	runtime := &Runtime{}
	runtime.Usage(os.Stderr)
	os.Exit(2)
}
func main() {
	logrus.SetOutput(os.Stderr)

	if len(os.Args) == 1 {
		Usage("")
	}
	keys := make([]string, 0, len(commands))
	for k := range commands {
		keys = append(keys, k)
	}
	index := -1
	var action string
	for v, value := range os.Args {
		if in(value, keys) {
			index = v
			action = value
			break
		}
	}
	runtime := &Runtime{}
	if index != -1 {
		var commonArgs []string
		if index > 1 {
			commonArgs = os.Args[1:index]
		} else {
			logrus.Debugf("No command args used\n")
		}
		logrus.Debugf("commonArgs %+v action: %s\n", commonArgs, action)
		if _, err := runtime.Parse(commonArgs); err != nil {
			panic(err)
		} else if action == "" {
			panic(errors.New("missing action"))
		} else if commands[action] == nil {
			panic(errors.New(fmt.Sprintf("'%s' command not defined", action)))
		}
		if runtime.debug {
			logrus.SetLevel(logrus.DebugLevel)
		}
		var executorArgs []string
		if len(os.Args) > (index + 1) {
			executorArgs = os.Args[index + 1:]
		}
		logrus.Debugf("executorArgs %+v\n", executorArgs)
		if action == "help" {
			Usage("your help:")
		} else if executor, err := commands[action].Parse(executorArgs); err != nil {
			logrus.Fatalf("%s failed because %+v\n", action, err)
		} else {
			if result, err2 := executor.Execute(runtime); err2 != nil {
				logrus.Fatalf("action %s execution failed because %+v\n", action, err2)
			} else {
				if bb, err7 := json.Marshal(result); err7 == nil {
					logrus.Infof("result %s\n", (string(bb)))
				}
			}
		}
	} else {
		if len(os.Args) > 1 {
			Usage("You need to include a verb")
		} else {
			Usage("Nothing to do.  You need to choose an action\n")
		}
	}
}