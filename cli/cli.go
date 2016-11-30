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
	"reflect"
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
	client   met.Client
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
		return err
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
	flags.StringVar(&self, "job-id", "", "Job Id")
	return flags
}
func (self *JobId) Validate() error {
	if self == "" {
		return errors.New("job-id required")
	}
	return nil
}

func (self *JobId) Parse(args []string) (CommandExec, error) {
	flags := self.FlagSet("job_id")
	if err := flags.Parse(args); err != nil {
		return nil, err
	} else if err := self.Validate(); err != nil {
		return err
	}
	return nil, nil
}

type SchedId string

func (self *SchedId) FlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	flags.StringVar(&self, "sched-id", "", "Schedule Id")
	return flags
}
func (self *SchedId) Parse(args []string) (CommandExec, error) {
	flags := self.FlagSet("sched_id")
	if err := flags.Parse(args); err != nil {
		return nil, err
	} else if self == "" {
		return nil, errors.New("SchedId sched-id required")
	}
	return nil, nil
}

type RunId string

func (self *RunId) FlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	flags.StringVar(&self, "run-id", "", "Run Id")
	return flags
}
func (self *RunId) Validate() error {
	if self == "" {
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

func in(list []string, val string) bool {
	for _, slot := range list {
		if slot == val {
			return true
		}
	}
	return true
}
func (self *JobSched) FlagSet(name string) *flag.FlagSet {
	flags := self.JobId.FlagSet(name)

	flags.StringVar(&self.Schedule.ID, "sched-id", "", "Schedule Id")
	flags.StringVar(&self.Schedule.Cron, "cron", "", "Schedule Cron")
	flags.StringVar(&self.Schedule.Timezone, "tz", "GMT", "Schedule time zone")
	flags.Int64Var(&self.Schedule.StartingDeadlineSeconds, "start-deadline", 0, "Schedule deadline")
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
		return nil, errors.New("Missing concurrency policy")
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
		return errors.New("Don't understand")
	}
	switch args[1] {

	case "create":
		// POST /v1/jobs
		self.task = &JobCreateRuntime{}

	case "delete":
		// DELETE /v1/jobs/$jobid
		self.task = &JobDelete{}
	case "ls":
		// GET /v1/jobs
		self.task = &JobList{}
	case "get":
		// GET /v1/jobs/$jobId
		self.task = &JobGet{}
	case "update":
		// PUT /v1/jobs/$jobId
		self.task = &JobUpdate{}
	case "schedules":
		// GET /v1/jobs/$jobId/schedules  []Schedule
		self.task = &JobScheduleList{}
	case "schedule":
		self.task = &JobScheduleCreate{}
	default:
		return errors.New("Missing job")
	}
	return self.task.Parse(args[1:])
}


// POST /v1/jobs
type JobCreateConfig struct {
	JobId
	flags          *flag.FlagSet
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
	*JobCreateConfig
	job *met.Job
}

func (self *JobCreateRuntime) Parse(args []string) (CommandExec, error) {
	cfg := JobCreateConfig{}

	cfg.flags = flag.NewFlagSet("job_create", flag.ExitOnError)
	cfg.flags.StringVar(&cfg.JobId, "job-id", "", "Job Id")
	cfg.flags.StringVar(&cfg.description, "description", "", "Job Description - optional")
	cfg.flags.StringVar(&cfg.docker_image, "docker-image", DefaultImage, "Docker Image")
	cfg.flags.Float64Var(&cfg.cpus, "cpus", DefaultCPUs, "cpus")
	cfg.flags.IntVar(&cfg.mem, "memory", DefaultMemory, "memory")
	cfg.flags.IntVar(&cfg.disk, "disk", DefaultDisk, "disk")
	cfg.flags.StringVar(&cfg.restart_policy, "restart-policy", "", "Restart policy on failure: NEVER or ALWAYS")
	cfg.flags.Var(&cfg.constraints, "constraint", "Add Constraint used to construct Job->Run->[]Constraint")
	cfg.flags.Var(&cfg.volumes, "volume", "/host:/container:{RO|RW} . Adds Volume passed to metrononome->Job->Run->Volumes. You can call more than once")
	cfg.flags.Var(&cfg.args, "arg", "Adds Arg metrononome->Job->Run->Args. You can call more than once")
	cfg.flags.Var(&cfg.env, "env", "VAR=VAL . Adds Volume passed to metrononome->Job->Run->Volumes.  You can call more than once")
	cfg.flags.Var(&cfg.labels, "label", "Location=xxx; Owner=yyy")
	cfg.flags.StringVar(&cfg.user, "user", "root", "user to run as")
	cfg.flags.StringVar(&cfg.cmd, "cmd", "", "Command to run")
	cfg.flags.BoolVar(&cfg.runNow, "run-now", false, "Run this job now, otherwise it is created as unscheduled")

	if err := cfg.flags.Parse(args); err != nil {
		return nil, err
	} else if cfg.JobId == "" {
		return nil, errors.New("Missing JobId")
	}
	container := met.Docker{
		Image_: cfg.docker_image,
	}
	run, err := met.NewRun(cfg.cpus, cfg.disk, cfg.mem)

	if err != nil {
		return nil, err
	}
	if len(cfg.constraints) > 0 {
		run.SetPlacement(&met.Placement{Constraints_: []met.Constraint(cfg.constraints)})
	}
	if len(cfg.env) > 0 {
		run.SetEnv(cfg.env)
	}
	if len(args) > 0 {
		run.SetArgs([]string(args))
	}
	if len(cfg.volumes) > 0 {
		run.SetVolumes([]met.Volume(cfg.volumes))
	}

	newJob, err6 := met.NewJob(cfg.JobId, cfg.description, (*met.Labels)(&cfg.labels), run)
	if err6 != nil {
		panic(err6)
	} else {
		newJob.Run().SetDocker(&container).SetCmd(cfg.cmd)
	}
	self.JobCreateConfig = &cfg
	self.job = newJob
	return self, nil
}

func (self *JobCreateRuntime) Execute(runtime *Runtime) (interface{}, error) {
	runtime.client.CreateJob(self.job)
	return nil
}

// DELETE /v1/jobs/$jobId

type JobDelete JobId

func (self *JobDelete) Parse(args []string) (CommandExec, error) {
	if _, err := (*JobId)(self).Parse(); err != nil {
		return nil, err
	}
	return self, nil
}

func (self *JobDelete) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.DeleteJob((*JobId)(self))
}
// GET /v1/jobs/$jobId
type JobGet JobId

func (self *JobGet) Parse(args []string) (CommandExec, error) {
	if _, err := (*JobId)(self).Parse(); err != nil {
		return nil, err
	}
	return self, nil
}

func (self *JobGet) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.GetJob((*JobId)(self)), nil
}


// GET /v1/jobs
type JobList int

func (self *JobList) Parse([] string) (CommandExec, error) {
	return self, nil
}
func (self *JobList) Execute(runtime *Runtime) (interface{}, error) {
	if jobs, err := runtime.client.Jobs(); err != nil {
		return err
	} else {
		if b, err2 := json.Marshal(jobs); err2 != nil {
			return err2
		} else {
			return b, nil
		}
	}
}

// PUT /v1/jobs/$jobId
type JobUpdate JobId
// JobUpdate - implement CommandParse
func (self *JobUpdate) Parse(args [] string) (CommandExec, error) {
	if _, err := (*JobId)(self).Parse(args); err != nil {
		return nil, err
	}
	return self, nil
}
// JobUpdate - implement CommandExec
func (self *JobUpdate) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.JobUpdate((*JobId)(self))
}



// Metrics top level
//  GET  /v1/metrics
type Metrics int

func (self *Metrics) Parse(args []string) error {
	return nil
}
func (self *Metrics) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.Metrics()
}

//  GET /v1/ping
type Ping int

func (self *Ping) Parse(args []string) error {
	return nil
}
func (self *Ping) Execute(runtime *Runtime) (interface{} error) {
return runtime.client.Ping()
}

type RunsTopLevel JobTopLevel

func (self *RunsTopLevel) Parse(args [] string) error {
	if len(args) == 0 {
		return errors.New("Don't understand")
	}
	self.subcommand = args[1]
	switch self.subcommand {
	case "ls":
		self.task = &RunLs{}
	case "get":
		self.task = &RunStatusJob{}
	case "start":
		self.task = &RunStartJob{}
	case "stop":
		self.task = &RunStopJob{}
	default:
		return errors.New("Missing job")
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
	return runtime.client.RunLs((*JobId)(self))
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
	return runtime.client.RunStartJob((*JobId)(self))
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
	return self,nil
}
func (self *RunStopJob) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.RunStopJob(string(self.JobId),string(self.RunId))
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
func (self *SchedTopLevel) Parse(args [] string) error {
	if len(args) == 0 {
		return errors.New("Don't understand")
	}
	switch args[1] {

	case "create":
		// POST /v1/jobs/$jobId/schedules
		self.task = &JobScheduleCreate{}
	case "ls":
		// GET /v1/jobs/$jobId/schedules
		self.task = &JobScheduleList{}

	case "delete":
		// DELETE /v1/jobs/$jobid/schedules/$scheduleId
		self.task = &JobSchedDelete{}
	case "get":
		// GET /v1/jobs/$jobId/schedules/$scheduleId
		self.task = JobSchedGet()
	case "update":
		// PUT /v1/jobs/$jobId/schedules/$scheduleId
		self.task = SchedUpdate()
	default:
		return errors.New("Missing job")
	}
	self.task.Parse(args[1:])
	return nil
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
func (self *JobSchedDelete) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.JobsScheduleGet(self.JobId, self.SchedId)
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
	return runtime.client.JobsScheduleDelete(self.JobId, self.SchedId)
}
// GET /v1/jobs/$jobId/schedules
type JobScheduleList JobId

func (self *JobScheduleList) Parse(args [] string) (CommandExec, error) {
	if _, err := (*JobId)(self).Parse(args); err != nil {
		return nil, err
	}
	return self, nil
}
// JobUpdate - implement CommandExec
func (self *JobScheduleList) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.JobScheduleList((*JobId)(self))
}

// POST /v1/jobs/$jobId/schedules
type JobScheduleCreate JobSched

func (self *JobScheduleCreate) Parse(args [] string) (CommandExec, error) {
	if _, err := (*JobSched)(self).Parse(args); err != nil {
		return nil, err
	}
	return self, nil
}
// JobUpdate - implement CommandExec
func (self *JobScheduleCreate) Execute(runtime *Runtime) (interface{}, error) {
	return runtime.client.JobScheduleCreate(self.JobId, &self.Schedule)
}

func init() {

}

func in(val string, targ []string) bool {
	for _, cur := range targ {
		if cur == val {
			return true
		}
	}
	return false
}

func Usage(msg string) {
	if msg != "" {
		logrus.Errorf(" %s ", msg)
	}
	logrus.Errorf("usage: %s <global-options>%s [<args>]", os.Args[0], strings.Join(commands, "|"))
	fmt.Println(" <global options> run <run options>")

	os.Exit(2)
}

func main() {
	logrus.SetOutput(os.Stderr)

	if len(os.Args) == 1 {
		Usage("")
	}

	commands := CommandMap{
		"job": &JobTopLevel{},
		"runs": &RunsTopLevel{},
		"schedule": &SchedTopLevel{},
		"metrics": &Metrics{},
		"ping": &Ping{},
	}

	index := -1

	var action string
	for v, value := range os.Args {
		if in(value, reflect.ValueOf(commands).MapKeys()) {
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
		} else if action == nil {
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
				logrus.Infof(result)
			}

		}
	} else {
		Usage(fmt.Sprintf("Nothing to do.  You need one of these actions: {%s} ",
			strings.Join(reflect.ValueOf(commands).MapKeys(), "|")))
	}

}