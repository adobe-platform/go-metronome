package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
)

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

type RunT struct {
	Artifacts_ []artifactT  `json:"artifacts,omitempty"`
	Cmd_       string      `json:"cmd,omitempty"`

	Args_           []string          `json:"args,omitempty"`
	Cpus_           float64           `json:"cpus"`
	Mem_            int               `json:"mem"`
	Disk_           int               `json:"disk"`
	Docker_         *dockerT           `json:"docker,omitempty"`
	Env_            map[string]string `json:"env,omitempty"`
	MaxLaunchDelay_ int               `json:"maxLaunchDelay"`
	Placement_      *PlacementT        `json:"placement,omitempty"`
	Restart_        *restartT          `json:"restart,omitempty"`
	User_           string            `json:"user,omitempty"`
	Volumes_        []volumeT         `json:"volumes,omitempty"`
}

type Run interface {
	Artifacts() []Artifact
	Cmd() string

	Args() []string
	Cpus() float64
	Mem() int
	Disk() int
	Docker() Docker
	Env() *map[string]string
	MaxLaunchDelay() int
	Placement() Placement
	Restart() Restart
	User() string
	Volumes() []Volume
}
func (self *RunT) String() string {
	rez := fmt.Sprint("cpus: %f disk: %d mem: %d\n",self.Cpus_,self.Disk_,self.Mem_)
	if self.Artifacts_ != nil && len(self.Artifacts_)>0 {
		b, _ := json.Marshal(self.Artifacts_)
		rez += fmt.Sprintf("Artifacts: %s\n", b)
	}
	return rez

}



func (self *RunT) Artifacts() []Artifact {
	con := make([]Artifact, len(self.Artifacts_))
	if self.Artifacts_ == nil {
		for i, v := range self.Artifacts_ {
			con[i] = &v
		}
	}
	return con
}
func (self *RunT) Cmd() string {
	return self.Cmd_
}

func (self *RunT) Args() [] string {
	return self.Args_
}

func (self *RunT) Cpus() float64 {
	return self.Cpus_
}

func (self *RunT) Mem() int {
	return self.Mem_
}

func (self *RunT) Disk() int {
	return self.Disk_
}

func (self *RunT) Docker() Docker {
	return self.Docker_
}

func (self *RunT) Env() *map[string]string {
	return &self.Env_
}

func (self *RunT) MaxLaunchDelay() int {
	return self.MaxLaunchDelay_
}

func (self *RunT) Placement() Placement {
	return self.Placement_
}

func (self *RunT) Restart() Restart {
	return self.Restart_
}

func (self *RunT) User() string {
	return self.User_
}
// make Volume ifc
func (self *RunT) Volumes() []Volume {
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
	if disk <= 0 {
		return nil, required("Run.disk")
	}
	if cpus <= 0.0 {
		return nil, required("Run.cpus")
	}
	vg:= RunT{
		Artifacts_: make([]artifactT, 0,10 ),
		Args_: make([]string, 0,0),
		Cpus_: cpus ,
		Mem_: mem,
		Disk_: disk,
		Docker_:  nil,
		Env_: make(map[string]string),
		MaxLaunchDelay_: 0,
		Placement_: nil,
		Restart_: nil,
		Volumes_: make([]volumeT,0,0),
		}

	return &vg,nil
}
		

type JobsT struct {
	Description string `json:"description"`
	ID          string `json:"id"`
	Labels      struct {
		Location string `json:"location"`
		Owner    string `json:"owner"`
	} `json:"labels"`
	Run RunT `json:"run"`
}

const data = `{
	"attribute": "jim gaffigan",
	"operator": "EQ",
	"value": "hot pockets"
}`

const data2 = `{
       "constraints" :[
        {
	"attribute": "jim gaffigan",
	"operator": "EQ",
	"value": "hot pockets"
	},
	{
	"attribute": "jim care",
	"operator": "EQ",
	"value": "foo bar"
	}
]}`
const data3 = `
{"description":"Example Application","id":"prod.example.app","labels":{"location":"olympus","owner":"zeus"},"run":{"artifacts":[{"uri":"http://foo.test.com/application.zip","extract":true,"executable":true,"cache":false}],"cmd":"nuke --dry --master local","args":["nuke","--dry","--master","local"],"cpus":1.5,"mem":32,"disk":128,"docker":{"image":"foo/bla:test"},"env":{"MON":"test","CONNECT":"direct"},"maxLaunchDelay":3600,"placement":{"constraints":[{"attribute":"rack","operator":"EQ","value":"rack-2"}]},"restart":{"activeDeadlineSeconds":120,"policy":"NEVER"},"user":"root","volumes":[{"containerPath":"!mnt/test","hostPath":"/etc/guest","mode":"RW"}]}}
`
const data4 = `
{"description":"Example Application","id":"prod.example.app","labels":{"location":"olympus","owner":"zeus"},"run":{"artifacts":[{"uri":"http://foo.test.com/application.zip","extract":true,"executable":true,"cache":false}],"cmd":"nuke --dry --master local","args":["nuke","--dry","--master","local"],"cpus":1.5,"mem":32,"disk":128,"docker":{"image":"foo/bla:test"},"env":{"MON":"test","CONNECT":"direct"},"maxLaunchDelay":3600,"placement":{"constraints":[{"attribute":"rack","operator":"EQ","value":"rack-2"}]},"restart":{"activeDeadlineSeconds":120,"policy":"NEVER"},"user":"root","volumes":[{"containerPath":"!/mnt/test","hostPath":"/etc/guest","mode":"RWW"}]}}
`
const data5 = `{"description":"Example Application","id":"prod.example.app","labels":{"location":"olympus","owner":"zeus"},"run":{"artifacts":[{"uri":"http://foo.test.com/application.zip","extract":true,"executable":true,"cache":false}],"cmd":"nuke --dry --master local","args":["nuke","--dry","--master","local"],"cpus":1.5,"mem":32,"disk":128,"docker":{"image":"foo/bla:test"},"env":{"MON":"test","CONNECT":"direct"},"maxLaunchDelay":3600,"placement":{"constraints":[{"attribute":"rack","operator":"EQ","value":"rack-2"}]},"restart":{"activeDeadlineSeconds":120,"policy":"NEVER"},"user":"root","volumes":[{"containerPath":"/mnt/test","hostPath":"/etc/guest","mode":"RW"}]}}`


type j6 struct {
	v7 []int
}
type foo interface {
	Bar() []int
}
func (self *j6) Bar() []int {
	return self.v7
}

func main() {
	fmt.Println("Hello, playground")
	a := constraintT{}
	err := json.Unmarshal([]byte(data), &a)
	if err != nil {
		log.Fatal("Unmarshal failed", err)
	}
	fmt.Println("foo %+v", a)

	var b PlacementT
	err2 := json.Unmarshal([]byte(data2), &b)
	if err2 != nil {
		log.Fatal("Unmarshal failed", err2)
	}
	fmt.Println("contraint array %+v", b)

	var c JobsT
	err3 := json.Unmarshal([]byte(data5), &c)
	if err3 != nil {
		log.Fatal("Unmarshal failed", err3)
	}
	fmt.Println("contraint array %+v", c)
	fmt.Println("c.Run.Placement.Constraints: %+v\nc.Run.Volumes %+v\n", c.Run.Volumes)
	if res1B, err := json.Marshal(c); err != nil {
		panic(err)
	} else {

		fmt.Println(data5)
		fmt.Println(string(res1B))
	}
	if v, err5 := NewRun(0.4, 1, 1); err5 != nil {
		panic(err5)
	} else {
		fmt.Printf("Run: %+v\n", v)

		fmt.Printf("marshal Run basic")
		if js, err := json.Marshal(v); err != nil {
			panic(err)
		} else {
			fmt.Printf("%s\n", js)
		}

		fmt.Printf("Now, append args to existing Run\n")
		args := v.Args()
		args = append(args,"foobar")

		fmt.Printf("modified args: %+v parent: %+v\n",args,v.Args())
		if js2, err := json.Marshal(v); err != nil {
			panic(err)
		} else {
			fmt.Printf(">%s\n", js2)
		}

	}

	var dd=j6{
		v7: make([]int, 0, 0),
	}
	ff := dd.Bar()

	for _,j:= range []int{7,6,5,4,3,2,1} {
		dd.v7 = append(dd.v7,j)
	}
	fmt.Printf("ifc %+v\n",ff)
	fmt.Printf("dd.v7 %+v\n",dd)

}

