package main

import (
	"fmt"
	//	"time"
	met "github.com/adobe-platform/go-metronome/metronome"
	"github.com/Sirupsen/logrus"
	"flag"
	"os"
	//	"strconv"
	"strings"
	"errors"
	"encoding/json"
)

// Command line defaults
const (
	DefaultHTTPAddr = "http://localhost:9000"
	DefaultImage = "alpine:3.4"
	DefaultCPUs = 0.2
	DefaultMemory = 128
	DefaultDisk = 128
)

type Runtime struct {
	httpAddr string
	flags    *flag.FlagSet
	debug    bool
	client   met.Metronome
}
type CommandExec interface {
	Execute(*Runtime) (interface{}, error)
}
type CommandParse interface {
	Parse([]string) (CommandExec, error)
}
type CommandMap map[string]CommandParse


// Command line parameters overrides for flag

type RunArgs []string
type NvList map[string]string
type ConstraintList [] met.Constraint
type VolumeList [] met.Volume
type LabelList  met.Labels

func (i *RunArgs) String() string {
	return fmt.Sprintf("%s", *i)
}
// The second method is Set(value string) error
func (i *RunArgs) Set(value string) error {
	fmt.Printf("%s\n", value)
	*i = append(*i, value)
	return nil
}
func (i *LabelList) String() string {
	return fmt.Sprintf("%s", *i)
}
// The second method is Set(value string) error
func (i *LabelList) Set(value string) error {
	fmt.Printf("%s\n", value)
	v := strings.Split(value, ";")

	lb := LabelList{}
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
	*i = lb
	return nil

}

func (i *NvList) String() string {
	return fmt.Sprintf("%s", *i)
}

// The second method is Set(value string) error
func (i *NvList) Set(value string) error {
	fmt.Printf("%s\n", value)
	nv := strings.Split(value, "=")
	if len(nv) != 2 {
		return errors.New("Environment vars should be NAME=VALUE")
	}
	map[string]string(*i)[nv[0]] = nv[1]
	return nil
}
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
// Global flags

func (self *Runtime) Parse(args []string) (CommandExec, error) {
	self.flags = flag.NewFlagSet("global", flag.ExitOnError)
	self.flags.StringVar(&self.httpAddr, "metronome-url", DefaultHTTPAddr, "Set the Metronome address")
	self.flags.BoolVar(&self.debug, "debug", false, "Turn on debug")

	if err := self.flags.Parse(args); err != nil {
		return nil,err
	}
	config := met.NewDefaultConfig()
	config.URL = self.httpAddr
	if client, err := met.NewClient(config); err != nil {
		return nil, err
	} else {
		self.client = client
	}
	// No exec returned
	return nil, nil
}

// base class used to parse args for many commands requiring `job-id` parsing
type JobId string

func (self *JobId) FlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	flags.StringVar((*string)(self), "job-id", "", "Job Id")
	return flags
}
func (self *JobId) Validate() error {
	if string(*self) == "" {
		return errors.New("job-id required")
	}
	return nil
}

func (self *JobId) Parse(args []string) (CommandExec, error) {
	flags := self.FlagSet("job_id")
	if err := flags.Parse(args); err != nil {
		return nil, err
	} else if err := self.Validate(); err != nil {
		return nil,err
	}
	return nil, nil
}

type SchedId string

func (self *SchedId) FlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	flags.StringVar((*string)(self), "sched-id", "", "Schedule Id")
	return flags
}
func (self *SchedId) Parse(args []string) (CommandExec, error) {
	flags := self.FlagSet("sched_id")
	if err := flags.Parse(args); err != nil {
		return nil, err
	} else if string(*self) == "" {
		return nil, errors.New("SchedId sched-id required")
	}
	return nil, nil
}

type RunId string

func (self *RunId) FlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	flags.StringVar((*string)(self), "run-id", "", "Run Id")
	return flags
}
func (self *RunId) Validate() error {
	if string(*self) == "" {
		return errors.New("run-id required")
	}
	return nil
}
func (self *RunId) Parse(args []string) (CommandExec, error) {
	flags := self.FlagSet("run_id")
	if err := flags.Parse(args); err != nil {
		return nil, err
	} else if err := self.Validate(); err != nil {
		return nil, err
	}
	return nil, nil
}

// JobSched is
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

func (self *JobSched) FlagSet(name string) *flag.FlagSet {
	flags := self.JobId.FlagSet(name)

	flags.StringVar(&self.Schedule.ID, "sched-id", "", "Schedule Id")
	flags.StringVar(&self.Schedule.Cron, "cron", "", "Schedule Cron")
	flags.StringVar(&self.Schedule.Timezone, "tz", "GMT", "Schedule time zone")
	flags.IntVar(&self.Schedule.StartingDeadlineSeconds, "start-deadline", 0, "Schedule deadline")
	flags.StringVar(&self.Schedule.ConcurrencyPolicy, "concurrency-policy", "ALLOW", "Schedule concurrency.  One of ALLOW,FORBID,REPLACE")
	return flags
}
func (self *JobSched) Validate() error {
	if self.JobId == "" {
		return errors.New("Missing JobId in JobScheduleCreate")
	} else if self.Schedule.ID == "" {
		return errors.New("Missing SchedId in JobScheduleCreate")
	} else if self.Schedule.Cron == "" {
		return errors.New("Missing Cron in JobScheduleCreate")
	} else if !in(self.Schedule.ConcurrencyPolicy, []string{"ALLOW", "FORBID", "REPLACE"}) {
		return errors.New("Missing concurrency policy")
	}
	return nil
}
func (self *JobSched) Parse(args []string) (CommandExec, error) {
	flags := self.FlagSet("job_sched")

	if err := flags.Parse(args); err != nil {
		return nil, err
	} else if err = self.Validate(); err != nil {
		return nil, err
	}
	return nil, nil
}
// jobs top level cli options
type JobTopLevel struct {
	subcommand string
	task       CommandParse
}

func (self *JobTopLevel) Parse(args [] string) (CommandExec, error) {
	if len(args) == 0 {
		return nil,errors.New("Don't understand")
	}
	switch args[1] {

	case "create":
		// POST /v1/jobs
		 x:=CommandParse(&JobCreateRuntime{})
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
	default:
		return nil, errors.New("Missing job")
	}
	return self.task.Parse(args[1:])
}


// POST /v1/jobs
type JobCreateConfig struct {
	JobId

	cpus           float64
	disk           int
	mem            int
	description    string
	docker_image   string
	restart_policy string
	constraints    ConstraintList
	volumes        VolumeList
	env            NvList
	labels         LabelList
	args           RunArgs
	cmd            string
	user           string
	runNow         bool
}
type JobCreateRuntime struct {
	JobCreateConfig
	job *met.Job
}

func (self *JobCreateRuntime) FlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ExitOnError)

	flags.StringVar((*string)(&self.JobId), "job-id", "", "Job Id")
	flags.StringVar(&self.description, "description", "", "Job Description - optional")
	flags.StringVar((*string)(&self.docker_image), "docker-image", DefaultImage, "Docker Image")
	flags.Float64Var(&self.cpus, "cpus", DefaultCPUs, "cpus")
	flags.IntVar(&self.mem, "memory", DefaultMemory, "memory")
	flags.IntVar(&self.disk, "disk", DefaultDisk, "disk")
	flags.StringVar(&self.restart_policy, "restart-policy", "", "Restart policy on failure: NEVER or ALWAYS")
	flags.Var(&self.constraints, "constraint", "Add Constraint used to construct Job->Run->[]Constraint")
	flags.Var(&self.volumes, "volume", "/host:/container:{RO|RW} . Adds Volume passed to metrononome->Job->Run->Volumes. You can call more than once")
	flags.Var(&self.args, "arg", "Adds Arg metrononome->Job->Run->Args. You can call more than once")
	flags.Var(&self.env, "env", "VAR=VAL . Adds Volume passed to metrononome->Job->Run->Volumes.  You can call more than once")
	flags.Var(&self.labels, "label", "Location=xxx; Owner=yyy")
	flags.StringVar(&self.user, "user", "root", "user to run as")
	flags.StringVar(&self.cmd, "cmd", "", "Command to run")
	flags.BoolVar(&self.runNow, "run-now", false, "Run this job now, otherwise it is created as unscheduled")
	return flags
}

func (self *JobCreateRuntime) Validate() error{
	if self.JobId == "" {
		return errors.New("Missing JobId")
	}
	return nil
}

func (self *JobCreateRuntime) Parse(args []string) (CommandExec, error) {

	flags := self.FlagSet("job_create")

	if err := flags.Parse(args); err != nil {
		return nil, err
	} else if err2:= self.Validate(); err2 != nil {
		return nil,err2
	}
	container := met.Docker{
		Image_: self.docker_image,
	}
	run, err := met.NewRun(self.cpus, self.disk, self.mem)

	if err != nil {
		return nil, err
	}
	if len(self.constraints) > 0 {
		run.SetPlacement(&met.Placement{Constraints_: []met.Constraint(self.constraints)})
	}
	if len(self.env) > 0 {
		run.SetEnv(self.env)
	}
	if len(args) > 0 {
		run.SetArgs([]string(args))
	}
	if len(self.volumes) > 0 {
		run.SetVolumes([]met.Volume(self.volumes))
	}

	newJob, err6 := met.NewJob(string(self.JobId), self.description, (*met.Labels)(&self.labels), run)
	if err6 != nil {
		panic(err6)
	} else {
		newJob.Run().SetDocker(&container).SetCmd(self.cmd)
	}

	self.job = newJob
	return self, nil
}

func (self *JobCreateRuntime) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.CreateJob(self.job)
}

// DELETE /v1/jobs/$jobId

type JobDelete JobId

func (self *JobDelete) Parse(args []string) (CommandExec, error) {
	if _, err := (*JobId)(self).Parse(args); err != nil {
		return nil, err
	}
	return self, nil
}

func (self *JobDelete) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.DeleteJob((string)(*self))
}
// GET /v1/jobs/$jobId
type JobGet JobId

func (self *JobGet) Parse(args []string) (CommandExec, error) {
	if _, err := (*JobId)(self).Parse(args); err != nil {
		return nil, err
	}
	return self, nil
}

func (self *JobGet) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.GetJob( string(*self))
}


// GET /v1/jobs
type JobList int

func (self *JobList) Parse([] string) (CommandExec, error) {
	return self, nil
}
func (self *JobList) Execute(runtime *Runtime) (interface{}, error) {
	if jobs, err := runtime.client.Jobs(); err != nil {
		return nil, err
	} else {
		if b, err2 := json.Marshal(jobs); err2 != nil {
			return nil,err2
		} else {
			return b, nil
		}
	}
}

// PUT /v1/jobs/$jobId
type JobUpdate JobCreateRuntime
// JobUpdate - implement CommandParse
func (self *JobUpdate) Parse(args [] string) (CommandExec, error) {
	if _, err := (*JobUpdate)(self).Parse(args); err != nil {
		return nil, err
	} else if err2:= (*JobUpdate)(self).Validate(); err2 !=nil {
		return nil,err
	}
	return self, nil
}
// JobUpdate - implement CommandExec
func (self *JobUpdate) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.JobUpdate(string(self.JobId),self.job)
}



// Metrics top level
//  GET  /v1/metrics
type Metrics int

func (self *Metrics) Parse(args []string) (CommandExec,error) {
	return self,nil
}
func (self *Metrics) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.Metrics()
}

//  GET /v1/ping
type Ping int

func (self *Ping) Parse(args []string) (CommandExec,error) {
	logrus.Debugf("Ping.Parse: %+v\n", args)
	return self,nil
}
func (self *Ping) Execute(runtime *Runtime) (interface{}, error) {
	logrus.Debugf("Ping.execute\n")
	if msg, err := runtime.client.Ping(); err != nil {
		return nil, err
	} else {
		return msg,nil
	}
}

type RunsTopLevel JobTopLevel

func (self *RunsTopLevel) Parse(args [] string) (CommandExec,error) {
	if len(args) == 0 {
		return nil,errors.New("Don't understand")
	}
	self.subcommand = args[1]
	switch self.subcommand {
	case "ls":
		self.task = CommandParse(new(RunLs))
	case "get":
		self.task = CommandParse(new(RunStatusJob))
	case "start":
		self.task = CommandParse(new(RunStartJob))
	case "stop":
		self.task = CommandParse(new(RunStopJob))
	default:
		return nil,errors.New("Missing job")
	}
	return self.task.Parse(args[1:])
}

// GET /v1/jobs/$jobId/runs
type RunLs JobId

func (self *RunLs) Parse(args []string) (CommandExec, error) {
	if _, err := (*JobId)(self).Parse(args); err != nil {
		return nil, err
	} else {
		return self, nil
	}
}
func (self *RunLs) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.RunLs(string(*self))
}
// POST /v1/jobs/$jobId/runs
type RunStartJob JobId

func (self *RunStartJob) Parse(args []string) (CommandExec, error) {
	if _, err := (*JobId)(self).Parse(args); err != nil {
		return nil, err
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

func (self *RunStatusJob) Parse(args []string) (CommandExec, error) {
	flagSet := self.JobId.FlagSet("runstatus_delete_jobid")
	runIdSet := self.RunId.FlagSet("runstatus_delete_schedid")
	runIdSet.VisitAll(func(flag *flag.Flag) {
		flagSet.Var(flag.Value, flag.Name, flag.Usage)
	})
	if err := flagSet.Parse(args); err != nil {
		return nil, err
	} else if err = self.JobId.Validate(); err != nil {
		return nil, err
	} else if err = self.RunId.Validate(); err != nil {
		return nil, err
	} else {
		return self, nil
	}
}
func (self *RunStatusJob) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.RunStatusJob(string(self.JobId), string(self.RunId))
}
// POST /v1/jobs/$jobId/runs/$runId/action/stop
type RunStopJob RunStatusJob

func (self *RunStopJob) Parse(args []string) (CommandExec, error) {
	if _, err := (*RunStatusJob)(self).Parse(args); err != nil {
		return nil, err
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

func NewSchedTopLevel() *JobTopLevel {
	top := &JobTopLevel{}
	return top
}
func (self *SchedTopLevel) Parse(args [] string) (CommandExec,error){
	if len(args) == 0 {
		return nil,errors.New("Don't understand")
	}
	switch args[1] {

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
	default:
		return nil,errors.New("Missing job")
	}
	return self.task.Parse(args[1:])

}

type JobSchedBase struct {
	JobId
	SchedId
}

func (self *JobSchedBase) Validate() error {
	if self.JobId == "" {
		return errors.New("Missing JobId in JobScheduleCreate")
	} else if self.SchedId == "" {
		return errors.New("Missing SchedId in JobScheduleCreate")
	}
	return nil
}
func (self *JobSchedBase) Parse(args []string) (CommandExec, error) {
	flagSet := self.JobId.FlagSet("sched_delete_jobid")
	schedIdSet := self.SchedId.FlagSet("sched_delete_schedid")
	schedIdSet.VisitAll(func(flag *flag.Flag) {
		flagSet.Var(flag.Value, flag.Name, flag.Usage)
	})
	if err := flagSet.Parse(args); err != nil {
		return nil, err
	} else if err = self.Validate(); err != nil {
		return nil, err
	}
	return nil, nil
}
// GET /v1/jobs/$jobId/schedules/$scheduleId
type JobSchedGet JobSchedBase

func (self *JobSchedGet) Parse(args []string) (CommandExec, error) {
	if _, err := (*JobSchedBase)(self).Parse(args); err != nil {
		return nil, err
	} else {
		return self, nil
	}
}
func (self *JobSchedGet) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.JobScheduleGet(string(self.JobId), string(self.SchedId))
}


// DELETE /v1/jobs/$jobId/schedules/$scheduleId
type JobSchedDelete JobSchedBase

func (self *JobSchedDelete) Parse(args []string) (CommandExec, error) {
	if _, err := (*JobSchedBase)(self).Parse(args); err != nil {
		return nil, err
	} else {
		return self, nil
	}
}
func (self *JobSchedDelete) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.JobScheduleDelete(string(self.JobId), string(self.SchedId))
}
// GET /v1/jobs/$jobId/schedules
type JobScheduleList JobId

func (self *JobScheduleList) Parse(args [] string) (CommandExec, error) {
	if _, err := (*JobId)(self).Parse(args); err != nil {
		return nil, err
	}
	return self, nil
}
// JobScheduleList - implement CommandExec
func (self *JobScheduleList) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.JobScheduleList(string(*self))
}

// POST /v1/jobs/$jobId/schedules
type JobScheduleCreate JobSched

func (self *JobScheduleCreate) Parse(args [] string) (CommandExec, error) {
	if _, err := (*JobSched)(self).Parse(args); err != nil {
		return nil, err
	}
	return self, nil
}
// JobScheduleCreate- implement CommandExec
func (self *JobScheduleCreate) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.JobScheduleCreate(string(self.JobId), &self.Schedule)
}

// PUT /v1/jobs/$jobId/schedules/$scheduleId
type JobSchedUpdate JobSched

func (self *JobSchedUpdate) Parse(args []string) (CommandExec, error) {
	if _, err := (*JobSched)(self).Parse(args); err != nil {
		return nil, err
	} else {
		return self, nil
	}
}
func (self *JobSchedUpdate) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.JobScheduleUpdate(string(self.JobId), string(self.Schedule.ID),&self.Schedule)
}

type Help int
func (self *Help) Parse(args[]string) (CommandExec,error){
	Usage("")
	// no-op; Usage exists
	return nil,nil
}
var commands CommandMap
func init() {
	commands = CommandMap{
		"job": CommandParse(new(JobTopLevel)),
		"runs": CommandParse(new(RunsTopLevel)),
		"schedule": CommandParse(new(SchedTopLevel)),
		"metrics": CommandParse(new(Metrics)),
		"ping": CommandParse(new(Ping)),
		"help": CommandParse(new(Help)),
	}

}

func Usage(msg string) {
	if msg != "" {
		logrus.Errorf(" %s ", msg)
	}
	logrus.Errorf("usage: %s <global-options> {%s} [<args>]", os.Args[0], strings.Join([]string {
		"job",
		"runs",
		"schedule",
		"metrics",
		"ping",
		"help",
	}, "|"))
	fmt.Println(" <global options> run <run options>")

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

	if index != -1 {
		commonArgs := os.Args[0:index]
		runtime := &Runtime{}
		if runtime.debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		if _, err := runtime.Parse(commonArgs); err != nil {
			panic(err)
		} else if action == "" {
			panic(errors.New("missing action"))
		} else if commands[action] == nil {
			panic(errors.New(fmt.Sprintf("'%s' command not defined", action)))
		}
		if executor, err := commands[action].Parse(os.Args[index:]); err != nil {
			logrus.Fatalf("%s failed because %+v\n", action, err)
		} else {
			if result, err2 := executor.Execute(runtime); err2 != nil {
				logrus.Fatalf("action %s execution failed because %+v\n", action, err2)
			} else {
				logrus.Infof("result %+s\n",result)
			}

		}
	} else {
		Usage("Nothing to do.  You need to choose an actions\n")
	}

}