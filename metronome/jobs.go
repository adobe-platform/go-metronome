package metronome

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"time"
//	"bytes"
	"encoding/json"
//	"io"
	"regexp"

)
///

var constraintViol = errors.New("Bad constraint.  Must be EQ,LIKE,UNLIKE")
var mountViol = errors.New("Mount point must designate RW,RO")
var containerPathViol = errors.New("Bad container path.  Must match `^/[^/].*$`")

func required(msg string) error {
	if len(msg) == 0 {
		return errors.New("Missing Required message")
	}
	return errors.New(fmt.Sprintf("%s is required by metronome api", msg))
}

type Artifact interface {
	Uri() string
	Executable() bool
	Extract() bool
	Cache() bool
}

// Jobs is a slice of jobs
type artifactT struct {
	Uri_        string `json:"uri"`
	Executable_ bool   `json:"executable"`
	Extract_    bool   `json:"extract"`
	Cache_      bool   `json:"cache"`
}

func (self *artifactT) Uri() string {
	return self.Uri_
}
func (self *artifactT) Executable() bool {
	return self.Executable_
}
func (self *artifactT) Extract() bool {
	return self.Extract_
}
func (self *artifactT) Cache() bool {
	return self.Cache_
}

type dockerT struct {
	Image_ string `json:"image"`
}
type Docker interface {
	Image() string
}

func NewDockerImage(image string) (Docker, error) {
	if len(image) == 0 {
		return nil, required("Docker.Image requires a value")
	}
	return &dockerT{Image_: image}, nil
}

func (self *dockerT) Image() string {
	return self.Image_
}

// constraint

type Operator int

const (
	EQ Operator = 1 + iota
	LIKE
	UNLIKE
)

var constraint_operators = [...]string{
	"EQ",
	"LIKE",
	"UNLIKE",
}

func (self Operator) String() string {
	return constraint_operators[int(self)-1]
}

func (self *Operator) UnmarshalJSON(raw []byte) error {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return err
	}
	switch s {
	case "EQ":
		*self = EQ
	case "LIKE":
		*self = LIKE
	case "UNLIKE":
		*self = UNLIKE
	default:

		return constraintViol
	}
	return nil
}
func (self *Operator) MarshalJSON() ([]byte, error) {
	s := self.String()

	return []byte(fmt.Sprintf("\"%s\"", s)), nil
}

type constraintT struct {
	Attribute_ string `json:"attribute"`
	// operator is EQ, LIKE,UNLIKE
	Operator_ Operator `json:"operator"`
	Value_    string   `json:"value"`
}

type Constraint interface {
	Attribute() string
	Operator() Operator
	Value() string
}

func NewContraint(attribute string, op Operator, value string) (Constraint, error) {
	if attribute == "" {
		return nil, required("Constraint.attribute")
	}
	return &constraintT{Attribute_: attribute, Operator_: op, Value_: value}, nil
}

func (self *constraintT) Attribute() string {
	return self.Attribute_
}
func (self *constraintT) Operator() Operator {
	return self.Operator_
}
func (self *constraintT) Value() string {
	return self.Value_
}

type PlacementT struct {
	Constraints_ []constraintT `json:"constraints,omitempty"`
}

type Placement interface {
	Constraints() ([]Constraint, error)
}

func (self *PlacementT) Constraints() ([]Constraint, error) {
	con := make([]Constraint, len(self.Constraints_))
	for i, v := range self.Constraints_ {
		con[i] = &v
	}
	return con, nil
}

// volumes

type MountMode int
type ContainerPath string

const (
	RO MountMode = 1 + iota
	RW
)

var mount_modes = [...]string{
	"RO",
	"RW",
}

func (self MountMode) String() string {
	return mount_modes[int(self)-1]
}
func (self *MountMode) MarshalJSON() ([]byte, error) {
	s := self.String()
	return []byte(fmt.Sprintf("\"%s\"", s)), nil
}
func (self *MountMode) UnmarshalJSON(raw []byte) error {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return err
	}
	switch s {
	case "RO":
		*self = RO
	case "RW":
		*self = RW
	default:
		return mountViol
	}
	return nil
}
func (self *ContainerPath) UnmarshalJSON(raw []byte) error {
	if _, err := regexp.MatchString("^/[^/].*$", string(raw)); err != nil {
		return containerPathViol
	}
	*self = ContainerPath(raw)
	return nil
}
func (self *ContainerPath) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%s", string(*self))), nil
}
func (self *ContainerPath) String() string {
	return string(*self)
}

func NewContainerPath(path string) (self *ContainerPath, err error) {
	if _, err = regexp.MatchString("^/[^/].*$", path); err != nil {
		return nil, err
	}
	vg := ContainerPath(path)

	return &vg, nil

}

type volumeT struct {
	// minlength 1; pattern: ^/[^/].*$
	ContainerPath_ ContainerPath `json:"containerPath"`
	HostPath_      string        `json:"hostPath"`
	// Values: RW,RO
	Mode_ MountMode `json:"mode"`
}

type Volume interface {
	ContainerPath() (ContainerPath, error)
	HostPath() (string, error)
	Mode() (MountMode, error)
}

func (self *volumeT) ContainerPath() (ContainerPath, error) {
	return self.ContainerPath_, nil
}
func (self *volumeT) HostPath() (string, error) {
	return self.HostPath_, nil
}
func (self *volumeT) Mode() (MountMode, error) {
	return self.Mode_, nil
}

func NewVolume(containerPath ContainerPath, hostPath string, mode MountMode) (Volume, *error) {
	return &volumeT{
		ContainerPath_: containerPath,
		HostPath_:      hostPath,
		Mode_:          mode,
	}, nil
}

type restartT struct {
	ActiveDeadlineSeconds_ int    `json:"activeDeadlineSeconds"`
	Policy_                string `json:"policy"`
}
type Restart interface {
	ActiveDeadlineSeconds() int
	Policy() string
}

func (self *restartT) ActiveDeadlineSeconds() int {
	return self.ActiveDeadlineSeconds_
}

func (self *restartT) Policy() string {
	return self.Policy_
}

func NewRestart(activeDeadlineSeconds int, policy string) (Restart, error) {
	if len(policy) == 0 {
		return nil, required("length(Restart.policy)>0")
	}
	return &restartT{ActiveDeadlineSeconds_: activeDeadlineSeconds, Policy_: policy}, nil
}

type runT struct {
	Artifacts_ []artifactT `json:"artifacts"`
	Cmd_       string      `json:"cmd"`

	Args_           []string          `json:"args"`
	Cpus_           float64           `json:"cpus"`
	Mem_            int               `json:"mem"`
	Disk_           int               `json:"disk"`
	Docker_         dockerT           `json:"docker"`
	Env_            map[string]string `json:"env"`
	MaxLaunchDelay_ int               `json:"maxLaunchDelay"`
	Placement_      PlacementT        `json:"placement"`
	Restart_        restartT          `json:"restart"`
	User_           string            `json:"user"`
	Volumes_        []volumeT         `json:"volumes"`
}

type Run interface {
	Artifacts() []Artifact
	Cmd() string

	Args() []string
	Cpus() float64
	Mem() int
	Disk() int
	Docker() Docker
	Env() map[string]string
	MaxLaunchDelay() int
	Placement() Placement
	Restart() Restart
	User() string
	Volumes() []Volume
}


func (self *runT) Artifacts() []Artifact {
	con := make([]Artifact, len(self.Artifacts_))
	for i, v := range self.Artifacts_ {
		con[i] = &v
	}
	return con
}

func (self *runT) User() string {
	return self.User_
}
// make Volume ifc
func (self *runT) Volumes() []Volume {
	con := make([]Volume, len(self.Volumes_))
	for i, v := range self.Volumes_ {
		con[i] = &v
	}
	return con
}



func NewRun(cpus float64, mem int, disk int ) (Run, error) {
	if mem <= 0 {
		return nil, required("Run.memory")
	}
	if dist\ <= 0 {
		return nil, required
	}
	vg := RunT{
	Artifacts_: make([]artifactT, 0,1 )
	//Cmd_       string
Args_: make([]string, 0, 0)
Cpus_:           float64           `json:"cpus"`
Mem_            int               `json:"mem"`
Disk_           int               `json:"disk"`
Docker_         dockerT           `json:"docker"`
Env_            map[string]string `json:"env"`
MaxLaunchDelay_ int               `json:"maxLaunchDelay"`
Placement_      PlacementT        `json:"placement"`
Restart_        restartT          `json:"restart"`
User_           string            `json:"user"`
Volumes_        []volumeT         `json:"volumes"`
}
return &vg
}

	type JobsT struct {
		Description string `json:"description"`
		ID          string `json:"id"`
		Labels      struct {
				    Location string `json:"location"`
				    Owner    string `json:"owner"`
			    } `json:"labels"`
		Run runT `json:"run"`
	}

/*

{
  "$schema": "http://json-schema.org/schema#",
  "definitions": {
    "pathType": {
      "type": "string",
      "pattern": "^([a-z0-9]([a-z0-9-]*[a-z0-9]+)*)([.][a-z0-9]([a-z0-9-]*[a-z0-9]+)*)*$",
      "minLength": 1
    }
  },
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "id": {
      "$ref": "#/definitions/pathType",
      "description": "Unique identifier for the job consisting of a series of names separated by dots. Each name must be at least 1 character and may only contain digits (`0-9`), dashes (`-`), and lowercase letters (`a-z`). The name may not begin or end with a dash."

    },
    "description": {
      "type": "string",
      "description": "A description of this job."
    },
    "labels": {
      "type": "object",
      "description": "Attaching metadata to jobs can be useful to expose additional information to other services, so we added the ability to place labels on jobs (for example, you could label jobs staging and production to mark services by their position in the pipeline).",
      "additionalProperties": {
        "type": "string"
      }
    },
    "run": {
      "type": "object",
      "additionalProperties": false,
      "description": "A run specification",
      "properties": {
        "args": {
          "items": {
            "type": "string"
          },
          "type": "array",
          "description": "An array of strings that represents an alternative mode of specifying the command to run. This was motivated by safe usage of containerizer features like a custom Docker ENTRYPOINT. Either `cmd` or `args` must be supplied. It is invalid to supply both `cmd` and `args` in the same job."
        },
        "artifacts": {
          "type": "array",
          "description": "Provided URIs are passed to Mesos fetcher module and resolved in runtime.",
          "items": {
            "type": "object",
            "additionalProperties": false,
            "properties": {
              "uri": {
                "type": "string",
                "description": "URI to be fetched by Mesos fetcher module"
              },
              "executable": {
                "type": "boolean",
                "description": "Set fetched artifact as executable"
              },
              "extract": {
                "type": "boolean",
                "description": "Extract fetched artifact if supported by Mesos fetcher module"
              },
              "cache": {
                "type": "boolean",
                "description": "Cache fetched artifact if supported by Mesos fetcher module"
              }
            },
            "required": [ "uri" ]
          }
        },
        "cmd": {
          "description": "The command that is executed.  This value is wrapped by Mesos via `/bin/sh -c ${job.cmd}`.  Either `cmd` or `args` must be supplied. It is invalid to supply both `cmd` and `args` in the same job.",
          "type": "string",
          "minLength": 1
        },
        "cpus": {
          "type": "number",
          "description": "The number of CPU shares this job needs per instance. This number does not have to be integer, but can be a fraction.",
          "minimum": 0.01
        },
        "disk": {
          "type": "number",
          "description": "How much disk space is needed for this job. This number does not have to be an integer, but can be a fraction.",
          "minimum": 0
        },
        "docker": {
          "type": "object",
          "additionalProperties": false,
          "properties": {
            "image": {
              "type": "string",
              "documentation": "The docker repository image name."
            }
          },
          "required": ["image"]
        },
        "env": {
          "type": "object",
          "patternProperties": {
            ".*": {
              "oneOf": [
                { "type": "string" }
              ]
            }
          }
        },
        "maxLaunchDelay": {
          "type": "integer",
          "minimum": 1,
          "description": "The number of seconds until the job needs to be running. If the deadline is reached without successfully running the job, the job is aborted."
        },
        "mem": {
          "type": "number",
          "description": "The amount of memory in MB that is needed for the job per instance.",
          "minimum": 32
        },
        "placement": {
          "type": "object",
          "additionalProperties": false,
          "properties": {
            "constraints": {
              "type": "array",
              "description": "The array of constraints to place this job.",
              "items": {
                "type": "object",
                "additionalProperties": false,
                "properties": {
                  "attribute": {
                    "type": "string",
                    "description": "The attribute name for this constraint."
                  },
                  "operator": {
                    "type": "string",
                    "description": "The operator for this constraint.",
                    "enum": ["EQ", "LIKE", "UNLIKE"]
                  },
                  "value": {
                    "type": "string",
                    "description": "The value for this constraint."
                  }
                },
                "required": ["attribute", "operator"]
              }
            }
          }
        },
        "user": {
          "type": "string",
          "description": "The user to use to run the tasks on the agent."
        },
        "restart": {
          "type": "object",
          "additionalProperties": false,
          "documentation": "Defines the behavior if a task fails",
          "properties": {
            "policy": {
              "type": "string",
              "documentation": "The policy to use if a job fails. NEVER will never try to relaunch a job. ON_FAILURE will try to start a job in case of failure.",
              "enum": ["NEVER", "ON_FAILURE"]
            },
            "activeDeadlineSeconds": {
              "type": "integer",
              "documentation": "If the job fails, how long should we try to restart the job. If no value is set, this means forever."
            }
          },
          "required": ["policy"]
        },
        "volumes": {
          "type": "array",
          "documentation": "The list of volumes for this job.",
          "items": {
            "type": "object",
            "additionalProperties": false,
            "documentation": "A volume definition for this job.",
            "properties": {
              "containerPath": {
                "type": "string",
                "description": "The path of the volume in the container",
                "minLength": 1,
                "pattern": "^/[^/].*$"
              },
              "hostPath": {
                "type": "string",
                "description": "The path of the volume on the host",
                "minLength": 1
              },
              "mode": {
                "type": "string",
                "description": "Possible values are RO for ReadOnly and RW for Read/Write",
                "enum": ["RO", "RW"]
              }
            },
            "required": ["containerPath", "hostPath", "mode"]
          }
        }
      },
      "required": ["cpus", "mem", "disk"]
    }
  },
  "required": ["id", "run"]
}
Example:

{
  "description": "Example Application",
  "id": "prod.example.app",
  "labels": {
    "location": "olympus",
    "owner": "zeus"
  },
  "run": {
    "artifacts": [
      {
        "uri": "http://foo.test.com/application.zip",
        "extract": true,
        "executable": true,
        "cache": false
      }
    ],
    "cmd": "nuke --dry --master local",
    "args": ["nuke", "--dry", "--master", "local"],
    "cpus": 1.5,
    "mem": 32,
    "disk": 128,
    "docker": {
      "image": "foo/bla:test"
    },
    "env": {
      "MON": "test",
      "CONNECT": "direct"
    },
    "maxLaunchDelay": 3600,
    "placement": {
      "constraints": [
        {
          "attribute": "rack",
          "operator": "EQ",
          "value": "rack-2"
        }
      ]
    },
    "restart": {
      "activeDeadlineSeconds": 120,
      "policy": "NEVER"
    },
    "user": "root",
    "volumes": [
      {
        "containerPath": "/mnt/test",
        "hostPath": "/etc/guest",
        "mode": "RW"
      }
    ]
  }
}


 */

/*
// A Job defines a chronos job
// https://github.com/mesos/chronos/blob/master/docs/docs/api.md#job-configuration
type chronos.Job struct {
	Args string `json:"args"`
	Artifacts string `json:"artifacts"`
	Cmd string `json:"cmd"`
	Cpus string `json:"cpus"`
	Disk string `json:"disk"`
	Docker string `json:"docker"`
	//Env string `json:"env"`
	Env   []map[string]string `json:"environmentVariables,omitempty"`

	MaxLaunchDelay string `json:"maxLaunchDelay"`
	Mem string `json:"mem"`
	Placement string `json:"placement"`
	Restart string `json:"restart"`
	User string `json:"user"`
	Volumes string `json:"volumes"`

	Name                   string              `json:"name"`
	Command                string              `json:"command"`
	Shell                  bool                `json:"shell,omitempty"`
	Epsilon                string              `json:"epsilon,omitempty"`
	Executor               string              `json:"executor,omitempty"`
	ExecutorFlags          string              `json:"executorFlags,omitempty"`
	Retries                int                 `json:"retries,omitempty"`
	Owner                  string              `json:"owner,omitempty"`
	OwnerName              string              `json:"ownerName,omitempty"`
	Description            string              `json:"description,omitempty"`
	Async                  bool                `json:"async,omitempty"`
	SuccessCount           int                 `json:"successCount,omitempty"`
	ErrorCount             int                 `json:"errorCount,omitempty"`
	LastSuccess            string              `json:"lastSuccess,omitempty"`
	LastError              string              `json:"lastError,omitempty"`
	CPUs                   float32             `json:"cpus,omitempty"`
	Disk                   float32             `json:"disk,omitempty"`
	Mem                    float32             `json:"mem,omitempty"`
	Disabled               bool                `json:"disabled,omitempty"`
	SoftError              bool                `json:"softError,omitempty"`
	DataProcessingJobType  bool                `json:"dataProcessingJobType,omitempty"`
	ErrorsSinceLastSuccess int                 `json:"errorsSinceLastSuccess,omitempty"`
	URIs                   []string            `json:"uris,omitempty"`
	EnvironmentVariables   []map[string]string `json:"environmentVariables,omitempty"`
	Arguments              []string            `json:"arguments,omitempty"`
	HighPriority           bool                `json:"highPriority,omitempty"`
	RunAsUser              string              `json:"runAsUser,omitempty"`
	Container              *Container          `json:"container,omitempty"`
	Schedule               string              `json:"schedule,omitempty"`
	ScheduleTimeZone       string              `json:"scheduleTimeZone,omitempty"`
	Constraints            []map[string]string `json:"constraints,omitempty"`
	Parents                []string            `json:"parents,omitempty"`
}
*/

// FormatSchedule will return a chronos schedule that can be used by the job
// See https://github.com/mesos/chronos/blob/master/docs/docs/api.md#adding-a-scheduled-job for details
// startTime (time.Time): when you want the job to start. A zero time instant means start immediately.
// interval (string): How often to run the job.
// reps (string): How many times to run the job.
func FormatSchedule(startTime time.Time, interval string, reps string) (string, error) {
	if err := validateInterval(interval); err != nil {
		return "", err
	}

	if err := validateReps(reps); err != nil {
		return "", err
	}

	schedule := fmt.Sprintf("%s/%s/%s", reps, formatTimeString(startTime), interval)

	return schedule, nil
}

// RunOnceNowSchedule will return a schedule that starts immediately, runs once,
// and runs every 2 minutes until successful
func RunOnceNowSchedule() string {
	return "R1//PT2M"
}

// Jobs gets all jobs that chronos knows about
func (client *Client) Jobs() (*Jobs, error) {
	jobs := new(Jobs)

	err := client.apiGet(MetronomeAPIJobs, nil, jobs)

	if err != nil {
		return nil, err
	}

	return jobs, nil
}

// DeleteJob will delete a chronos job
// name: The name of job you wish to delete
func (client *Client) DeleteJob(name string) error {
	return client.apiDelete(path.Join(MetronomeAPIJob, name), nil, nil)
}

// DeleteJobTasks will delete all tasks associated with a job.
// name: The name of the job whose tasks you wish to delete
func (client *Client) DeleteJobTasks(name string) error {
	return client.apiDelete(path.Join(MetronomeAPIKillJobTask, name), nil, nil)
}

// StartJob can manually start a job
// name: The name of the job to start
// args: A map of arguments to append to the job's command
func (client *Client) StartJob(name string, args map[string]string) error {
	return client.apiPut(path.Join(MetronomeAPIJob, name), args, nil)
}

// AddScheduledJob will add a scheduled job
// job: The job you would like to schedule
func (client *Client) AddScheduledJob(job *Job) error {
	return client.apiPost(MetrononeAPIAddScheduledJob, nil, job, nil)
}

// AddDependentJob will add a dependent job
func (client *Client) AddDependentJob(job *Job) error {
	return client.apiPost(MetronomeAPIAddDependentJob, nil, job, nil)
}

// RunOnceNowJob will add a scheduled job with a schedule generated by RunOnceNowSchedule
func (client *Client) RunOnceNowJob(job *Job) error {
	job.Schedule = RunOnceNowSchedule()
	job.Epsilon = "PT10M"
	return client.AddScheduledJob(job)
}

func validateReps(reps string) error {
	if strings.HasPrefix(reps, "R") {
		return nil
	}

	return errors.New("Repetitions string not formatted correctly")
}

func validateInterval(interval string) error {
	if strings.HasPrefix(interval, "P") {
		return nil
	}

	return errors.New("Interval string not formatted correctly")
}

func formatTimeString(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.Format(time.RFC3339Nano)
}
