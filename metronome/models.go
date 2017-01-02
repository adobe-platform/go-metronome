package metronome

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

var whitespaceRe = regexp.MustCompile(`\s+`)

var errConstraintViol = errors.New("Bad constraint.  Must be EQ,LIKE,UNLIKE")
var errMountViol = errors.New("Mount point must designate RW,RO")
var errContainerPathViol = errors.New("Bad container path.  Must match `^/[^/].*$`")

func required(msg string) error {
	if len(msg) == 0 {
		return errors.New("Missing Required message")
	}
	return fmt.Errorf("%s is required by metronome api", msg)
}


// Artifact - Metronome Artifact
type Artifact struct {
	URI        string `json:"uri"`
	Executable bool   `json:"executable"`
	Extract    bool   `json:"extract"`
	Cache      bool   `json:"cache"`
}
// GetURI - return string copy
func (theArtifact *Artifact) GetURI() string {
	return theArtifact.URI
}
// IsExecutable - is the artifact executable
func (theArtifact *Artifact) IsExecutable() bool {
	return theArtifact.Executable
}
// ShouldExtract - does the artifact need to be extracted
func (theArtifact *Artifact) ShouldExtract() bool {
	return theArtifact.Extract
}
// ShouldCache - should the artifact be cached
func (theArtifact *Artifact) ShouldCache() bool {
	return theArtifact.Cache
}
// Docker - metronome limited docker image
type Docker struct {
	Image string `json:"image"`
}

// NewDockerImage  - create a new image
func NewDockerImage(image string) (*Docker, error) {
	if len(image) == 0 {
		return nil, required("Docker.Image requires a value")
	}
	return &Docker{Image: image}, nil
}
// GetImage - the docker image
func (docker *Docker) GetImage() string {
	return docker.Image
}

// constraint support

// Operator - constrain operator values
type Operator int

const (
	// EQ - Valid Operator values
	EQ Operator = 1 + iota
	// LIKE - operator
	LIKE
	// UNLIKE - operator
	UNLIKE
)

var constraintOperators = [...]string{
	"EQ",
	"LIKE",
	"UNLIKE",
}
// String - string rep of operator
func (theOp *Operator) String() string {
	return constraintOperators[int(*theOp) - 1]
}
func decodeOperator(op string) (Operator, error) {
	switch op {
	case "EQ":
		return EQ, nil
	case "LIKE":
		return LIKE, nil
	case "UNLIKE":
		return UNLIKE, nil
	default:
		fmt.Printf("Operator.UnmarshallJSON - unknown value '%s'\n", op)
		return -1, errConstraintViol
	}

}
// UnmarshalJSON - hand json unmarshalling
func (theOp *Operator) UnmarshalJSON(raw []byte) error {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return err
	}
	op, err := decodeOperator(s)
	if err != nil {
		return err
	}
	*theOp = op

	return nil
}
// MarshalJSON - implement json interface for marshalling json
func (theOp *Operator) MarshalJSON() ([]byte, error) {
	s := theOp.String()
	return json.Marshal(s)
	//return []byte(fmt.Sprintf("\"%s\"", s)), nil
}
// Constraint - Metronome constraint
type Constraint struct {
	Attribute string `json:"attribute"`
	// operator is EQ, LIKE,UNLIKE
	Operator  Operator `json:"operator"`
	Value     string   `json:"value"`
}
// StrToConstraint - takes constraint as described in Metronome documentation
func StrToConstraint(cli string) (*Constraint, error) {
	args := whitespaceRe.Split(cli, -1)
	if len(args) != 3 {
		return nil, errors.New("Not enough constraint args `attribute` {EQ|LIKE} value")
	}
	op, err := decodeOperator(args[1])
	if err != nil {
		return nil, err
	}
	return NewConstraint(args[0], op, args[2])
}
// NewConstraint - create Metronome constraint
func NewConstraint(attribute string, op Operator, value string) (*Constraint, error) {
	if attribute == "" {
		return nil, required("Constraint.attribute")
	}
	return &Constraint{Attribute: attribute, Operator: op, Value: value}, nil
}
// GetAttribute - accessor
func (theConstraint *Constraint) GetAttribute() string {
	return theConstraint.Attribute
}
// GetOperator - accessor
func (theConstraint *Constraint) GetOperator() Operator {
	return theConstraint.Operator
}
// GetValue - accessor
func (theConstraint *Constraint) GetValue() string {
	return theConstraint.Value
}
// Placement - Metronome placement
type Placement struct {
	Constraints []Constraint `json:"constraints"`
}
// GetConstraints - return constraints
func (thePlacement *Placement) GetConstraints() ([]Constraint, error) {
	return thePlacement.Constraints, nil
}

//
// volume types
//

// MountMode - type for constraining json values to those spec'd in metronome api
type MountMode int

// ContainerPath - type for contraining json values to limited paths spec'd in metronome api
type ContainerPath string

const (
	// RO - read-only mount
	RO MountMode = 1 + iota
	// RW - read/write mount
	RW
)

var mountModes = [...]string{
	"RO",
	"RW",
}
// String - stringrep
func (mm MountMode) String() string {
	return mountModes[int(mm) - 1]
}

// MarshalJSON - json interface requirement
func (mm *MountMode) MarshalJSON() ([]byte, error) {
	s := mm.String()
	return json.Marshal(s)
}

func decodeMount(mode string) (MountMode, error) {
	switch mode {
	case "RO":
		return RO, nil
	case "RW":
		return RW, nil
	default:
		return -1, errMountViol
	}

}
// UnmarshalJSON - json interface implementation. Ensure we have a valid mount value
func (mm *MountMode) UnmarshalJSON(raw []byte) error {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return err
	}
	mode, err := decodeMount(s)
	if err != nil {
		return err
	}
	*mm = mode

	return nil
}
// MarshalJSON - json interface implementation
func (cp *ContainerPath) MarshalJSON() ([]byte, error) {
	s := string(*cp)
	return json.Marshal(s)
}
// UnmarshalJSON - json interface implementation.  ensure valid path as spec'd by metronome api
func (cp *ContainerPath) UnmarshalJSON(raw []byte) error {

	// byte must be unmarshalled as a string otherwise there are cases where the quotes will bleed
	// through
	var s string
	json.Unmarshal(raw, &s)
	if _, err := regexp.MatchString("^/[^/].*$", s); err != nil {
		return errContainerPathViol
	}

	*cp = ContainerPath(s)
	return nil
}
// NewContainerPath - create a new container path that's checked for validity per Metronome's doc
func NewContainerPath(path string) (self ContainerPath, err error) {
	if _, err = regexp.MatchString("^/[^/].*$", path); err != nil {
		return "", err
	}
	vg := ContainerPath(path)

	return vg, nil

}
// Volume - structure representing a Metronome Volume validated types
type Volume struct {
	// minlength 1; pattern: ^/[^/].*$
	ContainerPath ContainerPath `json:"containerPath"`
	HostPath      string        `json:"hostPath"`
	// Values: RW,RO
	Mode          MountMode `json:"mode"`
}
// NewVolume - creates a new volume from raw strings
func NewVolume(rawPath string, hostPath string, modestr string) (*Volume, error) {
	mode, err := decodeMount(modestr);
	if err != nil {
		return nil, err
	}
	vol := Volume{Mode: mode, HostPath:hostPath}
	// ensure valid path
	cpath, err := NewContainerPath(rawPath);
	if err != nil {
		return nil, err
	}
	vol.ContainerPath = cpath

	if vol.HostPath == "" {
		return nil, required("host path")
	}
	return &vol, nil

}
// Restart - structure representing a Metronome structure
type Restart struct {
	ActiveDeadlineSeconds int    `json:"activeDeadlineSeconds"`
	Policy                string `json:"policy"`
}
// NewRestart - create a valid Restart policy
func NewRestart(activeDeadlineSeconds int, policy string) (*Restart, error) {
	if len(policy) == 0 {
		return nil, required("length(Restart.policy)>0")
	} else if !(policy == "NEVER" || policy == "ON_FAILURE") {
		return nil, fmt.Errorf("Policy must be 'ON_FAILURE' or 'NEVER' not %s", policy)
	}
	return &Restart{ActiveDeadlineSeconds: activeDeadlineSeconds, Policy: policy}, nil
}
// Run - composite structure representing Metronone run
type Run struct {
	Artifacts      []Artifact  `json:"artifacts,omitempty"`
	Cmd            string      `json:"cmd,omitempty"`

	Args           []string          `json:"args,omitempty"`
	Cpus           float64           `json:"cpus"`
	Mem            int               `json:"mem"`
	Disk           int               `json:"disk"`
	Docker         *Docker           `json:"docker,omitempty"`
	Env            map[string]string `json:"env,omitempty"`
	MaxLaunchDelay int               `json:"maxLaunchDelay"`
	Placement      *Placement        `json:"placement,omitempty"`
	Restart        *Restart          `json:"restart,omitempty"`
	User           string            `json:"user,omitempty"`
	Volumes        []Volume         `json:"volumes"`
}

// GetArtifacts - accessor returning Artifacts
func (runner *Run) GetArtifacts() []Artifact {
	return runner.Artifacts
}
// SetArtifacts - set Artifacts returning pointer to Run so they can be setters can be chained together
func (runner *Run) SetArtifacts(artifacts []Artifact) *Run {
	runner.Artifacts = artifacts
	return runner
}
// GetCmd - accessor for the cmd to run
func (runner *Run) GetCmd() string {
	return runner.Cmd
}
// SetCmd - set the command to run
func (runner *Run) SetCmd(cmd string) *Run {
	runner.Cmd = cmd
	return runner
}
// GetArgs - accessor returning the list of arguments
func (runner *Run) GetArgs() *[]string {
	return &runner.Args
}
// AddArg - appends an argument to the list
func (runner *Run) AddArg(item string) {
	runner.Args = append(runner.Args, item)
}
// SetArgs - replace the entire argument list
func (runner *Run) SetArgs(newargs [] string) *Run {
	runner.Args = newargs
	return runner
}
// GetCpus - the number of cpus to assign
func (runner *Run) GetCpus() float64 {
	return runner.Cpus
}
// SetCpus - set the number of cpus to use
func (runner *Run) SetCpus(p float64) *Run {
	runner.Cpus = p
	return runner
}
// GetMem - accessor returning the amount of memory to use
func (runner *Run) GetMem() int {
	return runner.Mem
}
// SetMem - set the amount of memory to assign
func (runner *Run) SetMem(p int) *Run {
	runner.Mem = p
	return runner
}
// GetDisk  - accesor returning the disk space to assign
func (runner *Run) GetDisk() int {
	return runner.Disk
}
// SetDisk - set the amount of disk space to use
func (runner *Run) SetDisk(p int) *Run {
	runner.Disk = p
	return runner
}
// GetDocker - accessor returning the docker structure if set
func (runner *Run) GetDocker() *Docker {
	return runner.Docker
}
// SetDocker - replace the Docker image to use
func (runner *Run) SetDocker(docker *Docker) *Run {
	runner.Docker = docker
	return runner
}
// GetEnv - return the current environment
func (runner *Run) GetEnv() map[string]string {
	return runner.Env
}
// SetEnv - replace the environment to use
func (runner *Run) SetEnv(mp map[string]string) *Run {
	runner.Env = mp
	return runner
}
// GetMaxLaunchDelay - accessor returning the maximum launch delay
func (runner *Run) GetMaxLaunchDelay() int {
	return runner.MaxLaunchDelay
}
// SetMaxLaunchDelay - set the delay
func (runner *Run) SetMaxLaunchDelay(p int) *Run {
	runner.MaxLaunchDelay = p
	return runner
}
// GetPlacement - get the placement
func (runner *Run) GetPlacement() *Placement {
	return runner.Placement
}
// SetPlacement - set the job placement
func (runner *Run) SetPlacement(p *Placement) *Run {
	runner.Placement = p
	return runner
}
// GetRestart - the restart structure
func (runner *Run) GetRestart() *Restart {
	return runner.Restart
}
// SetRestart - set the restart structure
func (runner *Run) SetRestart(restart *Restart) *Run {
	runner.Restart = restart
	return runner
}
// GetUser - get the user
func (runner *Run) GetUser() string {
	return runner.User
}
// SetUser - set the user
func (runner *Run) SetUser(user string) *Run {
	runner.User = user
	return runner
}

// GetVolumes - get job's Volume mappings
func (runner *Run) GetVolumes() *[]Volume {
	return &runner.Volumes
}
// SetVolumes - set the job's Volume mappings
func (runner *Run) SetVolumes(vols []Volume) *Run {
	runner.Volumes = vols
	return runner
}
// NewRun - create a run structure needed for a job
func NewRun(cpus float64, mem int, disk int) (*Run, error) {
	if mem <= 0 {
		return nil, required("Run.memory")
	}
	if disk <= 0 {
		return nil, required("Run.disk")
	}
	if cpus <= 0.0 {
		return nil, required("Run.cpus")
	}
	vg := Run{
		Artifacts: make([]Artifact, 0, 10),
		Args: make([]string, 0, 0),
		Cpus: cpus,
		Mem: mem,
		Disk: disk,
		Docker:  nil,
		Env: make(map[string]string),
		MaxLaunchDelay: 0,
		Placement: nil,
		Restart: nil,
		Volumes: make([]Volume, 0, 0),
	}
	return &vg, nil
}
// Labels - list of labels that get converted to environment variables on job
type Labels map[string]string

//Job - toplevel metronome structure for creating and managing a job
type Job struct {
	Description    string `json:"description"`
	ID             string `json:"id"`
	Labels         *Labels`json:"labels,omitempty"`
	Run            *Run `json:"run"`
	Schedules      []*Schedule `json:"schedules,omitempty"`
	ActiveRuns     [] *ActiveRun`json:"activeRuns,omitempty"`
	History        *History `json:"history,omitempty"`
	HistorySummary *HistorySummary `json:"historySummary,omitempty"`
}
//NewJob - create a job checking for some required fields
func NewJob(id string, description string, labels Labels, run *Run) (*Job, error) {

	if len(id) == 0 {
		return nil, required("Job.Id")
	}
	if run == nil {
		return nil, required("Job.run")
	}

	return &Job{ID: id,
		Description: description,
		Labels: &labels,
		Run: run,
	}, nil
}
// The following methods only effect a local job structure.  To apply to an existing metronome job, update job must be called

// GetID - get the job id
func (theJob *Job) GetID() string {
	return theJob.ID
}
// SetID - the job id
func (theJob *Job) SetID(id string) *Job {
	theJob.ID = id
	return theJob
}
// GetDescription - get the job description
func (theJob *Job) GetDescription() string {
	return theJob.Description
}
// SetDescription - set the job description.
func (theJob *Job) SetDescription(desc string) *Job {
	theJob.Description = desc
	return theJob
}
// GetRun - the Run structure for the job
func (theJob *Job) GetRun() *Run {
	return theJob.Run
}
// SetRun - set the run structure
func (theJob *Job) SetRun(run *Run) *Job {
	theJob.Run = run
	return theJob
}
// GetLabels - get the job labels
func (theJob *Job) GetLabels() *Labels {
	return theJob.Labels
}
// SetLabel - set the job lables
func (theJob *Job) SetLabel(label Labels) *Job {
	theJob.Labels = &label
	return theJob
}
// Schedule - represent a metronome schedule
type Schedule struct {
	ID                      string `json:"id"`
	Cron                    string `json:"cron"`
	ConcurrencyPolicy       string `json:"concurrencyPolicy"`
	Enabled                 bool `json:"enabled"`
	StartingDeadlineSeconds int `json:"startingDeadlineSeconds"`
	Timezone                string `json:"timezone"`
	NextRunAt               string `json:"nextRunAt,omitempty"`
}
// JobStatus - represents a metronome job status
type JobStatus struct {
	CompletedAt interface{} `json:"completedAt"`
	CreatedAt   string `json:"createdAt"`
	ID          string `json:"id"`
	JobID       string `json:"jobId"`
	Status      string `json:"status"`
	Tasks       [] TaskStatus `json:"tasks"`
}
// Jobs - list of jobs
type Jobs []Job

// TaskStatus - status of currently running task representing job
type TaskStatus struct {
	ID        string `json:"id"`
	StartedAt string `json:"startedAt"`
	Status    string `json:"status"`
}
// HistoryStatus - history outcome of previous jobs
type HistoryStatus struct {
	ID         string `json:"id"`
	CreatedAt  string `json:"createdAt"`
	FinishedAt string `json:"finishedAt"`
}
// ActiveRun - undocumented structure returned via api for job runs
type ActiveRun struct {
	ID          string `json:"id"`
	JobID       string `json:"jobId"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	CompletedAt interface{} `json:"completedAt"`
	Tasks       []TaskStatus `json:"tasks"`
}
// History - undocumented structure returned by Metronome api for job runs
type History struct {
	SuccessCount           int `json:"successCount"`
	FailureCount           int `json:"failureCount"`
	LastSuccessAt          string `json:"lastSuccessAt"`
	LastFailureAt          string `json:"lastFailureAt"`
	SuccessfulFinishedRuns [] HistoryStatus `json:"successfulFinishedRuns"`
	FailedFinishedRuns     [] HistoryStatus `json:"failedFinishedRuns"`
}
// HistorySummary - undocumented structure returned by Metronome api for job runs
type HistorySummary struct {
	SuccessCount  int `json:"successCount"`
	FailureCount  int `json:"failureCount"`
	LastSuccessAt string `json:"lastSuccessAt"`
	LastFailureAt string `json:"lastFailureAt"`
}