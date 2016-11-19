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

/*
const data = `{
	"attribute": "jim gaffigan",
	"operator": "EQ",
	"value": "hot pockets"
}`

const data2 =`{
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
	}

*/
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
	return []byte(fmt.Sprintf("\"%s\"",s)), nil
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
func (self *ContainerPath) MarshalJSON() ([]byte, error) {
	s := ContainerPath(*self)
	return []byte(fmt.Sprintf("\"%s\"",s)), nil
}

func (self *ContainerPath) UnmarshalJSON(raw []byte) error {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return err
	}
	if _, err := regexp.MatchString("^/[^/].*$", s); err != nil {
		return err
	}
	fmt.Printf("Validated ContainerPath %+v\n", s)
	self = &ContainerPath{s}
//	fmt.Printf("Validated ContainerPath2 %+v\n", string(vg))
	return nil
}
func (self ContainerPath) String() string {

	return string(self)
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

type Jobs struct {
	Description string `json:"description"`
	ID          string `json:"id"`
	Labels      struct {
		Location string `json:"location"`
		Owner    string `json:"owner"`
	} `json:"labels"`
	Run struct {
		Artifacts []artifactT `json:"artifacts"`
		Cmd       string      `json:"cmd"`

		Args   []string `json:"args"`
		Cpus   float64  `json:"cpus"`
		Mem    int      `json:"mem"`
		Disk   int      `json:"disk"`
		Docker struct {
			Image string `json:"image"`
		} `json:"docker"`
		Env            map[string]string `json:"env"`
		MaxLaunchDelay int               `json:"maxLaunchDelay"`
		Placement      PlacementT        `json:"placement"`
		Restart        struct {
			ActiveDeadlineSeconds int    `json:"activeDeadlineSeconds"`
			Policy                string `json:"policy"`
		} `json:"restart"`
		User    string    `json:"user"`
		Volumes []volumeT `json:"volumes"`
	} `json:"run"`
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

	var c Jobs
	err3 := json.Unmarshal([]byte(data5), &c)
	if err3 != nil {
		log.Fatal("Unmarshal failed", err3)
	}
	fmt.Println("contraint array %+v", c)
	fmt.Println("c.Run.Placement.Constraints: %+v\nc.Run.Volumes %+v\n", c.Run.Volumes[0].ContainerPath_)
	if res1B, err := json.Marshal(c); err != nil {
		panic(err)
	} else {

		fmt.Println(data5)
		fmt.Println(string(res1B))
	}
}
