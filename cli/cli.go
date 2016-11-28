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
)

// Command line defaults
const (
	DefaultHTTPAddr = ":11000"
	DefaultImage = "libmesos/ubuntu"
	DefaultCPUs = 0.2
	DefaultMemory = 128
	DefaultDisk = 128
)

// Command line parameters
var httpAddr string
var image string
var debug bool
var now bool = false

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
	for _,ii := range v {
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
	if lb.Location ==""  && lb.Owner == ""{

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

var (
	globalFlagSet *flag.FlagSet
	job_create *flag.FlagSet
	chronosFlagSet *flag.FlagSet

	scheduleFlagSet *flag.FlagSet

	cpus float64
	disk int
	mem int
	description string
	docker_image string
	restart_policy string
	chronos_sched string
	metronome_sched string
	constraints ConstraintList
	volumes VolumeList
	env NvList
	labels LabelList
	args RunArgs
	cmd string
	subcommands []string
	user string
	jobId string
	cronTab string
)

func init() {
	subcommands = []string{
		"job",
		"run",
		"schedule",
		"chronos",
	}
	globalFlagSet = flag.NewFlagSet("common", flag.ExitOnError)
	globalFlagSet.StringVar(&httpAddr, "metronome-url", DefaultHTTPAddr, "Set the Metronome address")
	globalFlagSet.BoolVar(&debug, "debug", false, "Turn on debug")

	job_create = flag.NewFlagSet("job_create", flag.ExitOnError)
	job_create.StringVar(&jobId, "jobId", "", "Job Id - require")
	job_create.StringVar(&description, "description", "", "Job Description - optional")
	job_create.StringVar(&image, "docker-image", DefaultImage, "Docker Image")
	job_create.Float64Var(&cpus, "cpus", DefaultCPUs, "cpus")
	job_create.IntVar(&mem, "memory", DefaultMemory, "memory")
	job_create.IntVar(&disk, "disk", DefaultDisk, "disk")
	job_create.StringVar(&docker_image, "docker-image", "", "docker image to run")
	job_create.StringVar(&restart_policy, "restart-policy", "", "Restart policy on failure: NEVER or ALWAYS")
	job_create.Var(&constraints, "constraint", "Add Constraint used to construct Job->Run->[]Constraint")
	job_create.Var(&volumes, "volume", "/host:/container:{RO|RW} . Adds Volume passed to metrononome->Job->Run->Volumes. You can call more than once")
	job_create.Var(&args, "arg", "Adds Arg metrononome->Job->Run->Args. You can call more than once")
	job_create.Var(&env, "env", "VAR=VAL . Adds Volume passed to metrononome->Job->Run->Volumes.  You can call more than once")
	job_create.Var(&labels, "label", "Location=xxx; Owner=yyy")
	job_create.StringVar(&user, "user", "root", "user to run as")
	job_create.StringVar(&cmd, "cmd", "", "Command to run")
	job_create.BoolVar(&now, "run-now", false, "Run this job now, otherwise it is created as unscheduled")

	scheduleFlagSet = flag.NewFlagSet("schedule", flag.ExitOnError)
	scheduleFlagSet.StringVar(&jobId, "job-id", "", "Job Id")
	scheduleFlagSet.StringVar(&cronTab, "crontab", "", "Metronome/cron schedule  ex. */2 * * * * *")

	chronosFlagSet = flag.NewFlagSet("chronos", flag.ExitOnError)
	chronosFlagSet.StringVar(&chronos_sched, "iso8601-schedule", "", "Run later using chronos schedule")

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
	logrus.Errorf("usage: %s <global-options>%s [<args>]", os.Args[0], strings.Join(subcommands, "|"))
	fmt.Println(" <global options> run <run options>")

	os.Exit(2)
}

func main() {
	logrus.SetOutput(os.Stderr)

	if len(os.Args) == 1 {
		Usage("")
	}

	index := -1
	for v, value := range os.Args {
		if in(value, subcommands) {
			index = v
			break
		}
	}
	var action, subaction string

	if index != -1 {
		commonArgs := os.Args[0:index]
		globalFlagSet.Parse(commonArgs)

		action = os.Args[index]
		switch action {
		case "job":
			// subcommand
			subaction = os.Args[index + 1]
			switch subaction {
			case "create":
				if err := job_create.Parse(os.Args[index + 2:]); err != nil {
					job_create.Usage()
				} else if jobId == "" {
					Usage("Missing JobId")
				} else if cmd == "" {
					Usage("Missing Command")
				}

			case "delete":
			case "ls":

			case "run":

			}
		case "schedule":
			if len(os.Args) > (index + 2) && os.Args[index + 1] == "chronos" {
				chronosFlagSet.Parse(os.Args[index + 2:])
			} else if len(os.Args) > (index + 1) {
				scheduleFlagSet.Parse(os.Args[index + 1:])
			} else {
				Usage("Not enough arguments to schedule")
			}
		default:
			Usage(fmt.Sprintf("%q is not valid command.\n", os.Args[1]))
		}
	} else {
		Usage("Nothing to do")
	}
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	config := met.NewDefaultConfig()
	config.URL = httpAddr
	client, err := met.NewClient(config)

	if err != nil {
		fmt.Println("Could not create client: ", err)
		return
	}
	switch action {
	case "job":
		switch subaction {
		case "ls":
			if jobs, err := client.Jobs(); err != nil {
				panic(err)
			} else {
				logrus.Debug(jobs)
			}
		case "create":
			// Add a scheduled job
			//	cronstr, _ := met.FormatSchedule(*new(time.Time), "PT2M", "R1")

			container := met.Docker{
				Image_: image,
			}
			run, err := met.NewRun(cpus, disk, mem)

			if err != nil {
				panic(err)
			}
			if len(constraints) > 0 {
				run.SetPlacement(&met.Placement{Constraints_: []met.Constraint(constraints)})
			}
			if len(env) > 0 {
				run.SetEnv(env)
			}
			if len(args) > 0 {
				run.SetArgs([]string(args))
			}
			if len(volumes) > 0 {
				run.SetVolumes([]met.Volume(volumes))
			}

			newJob, err6 := met.NewJob(jobId, description,(* met.Labels)(&labels), run)
			if err6 != nil {
				panic(err6)
			} else {
				newJob.Run().SetDocker(&container).SetCmd("echo 'Hello world'")
			}


		}

	}
	/*
	sched := &met.Schedule{
		Cron: cronstr,

	}
	client.AddScheduledJob(&newJob, &sched)

	// Get all current jobs
	jobs, err := client.Jobs()
	fmt.Println("Current jobs:")
	for _, job := range *jobs {
		fmt.Println("Job Id: ", job.ID_)
	}

	// Delete the job
	client.DeleteJob("my.test.job")

	// Get all current jobs
	jobs, _ = client.Jobs()
	fmt.Println("Current jobs:")
	for _, job := range *jobs {
		fmt.Println("Job Name: ", job.ID_)
	}

	// To run a job immediately, and only once
	oneTimeJob := newJob.SetId("myOneTimeJob")
	client.RunOnceNowJob(&oneTimeJob)
	client.DeleteJob("myOneTimeJob")
	*/
}
