package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
)

var mountViol = errors.New("Mount point must designate RW,RO")
var containerPathViol = errors.New("Bad container path.  Must match `^/[^/].*$`")

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
func (self *ContainerPath) MarshalJSON() ([]byte, error) {
	s := ContainerPath(*self)
	return []byte(fmt.Sprintf("\"%s\"", s)), nil
}

func (self *ContainerPath) UnmarshalJSON(raw []byte) error {
	if _, err := regexp.MatchString("^/[^/].*$", string(raw)); err != nil {
		return err
	}
	fmt.Printf("Validated ContainerPath %+v\n", string(raw))
	*self = ContainerPath(raw)


	return nil
}
func (self ContainerPath) String() string {
	fmt.Printf("ContainerPath string called\n")
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

const data = `
{"containerPath":"/mnt/test","hostPath":"/etc/guest","mode":"RW"}
`

func main() {
	fmt.Println("Hello, playground")
	a := volumeT{}
	err := json.Unmarshal([]byte(data), &a)
	if err != nil {
		log.Fatal("Unmarshal failed", err)
	}
	fmt.Println("foo %+v", a.ContainerPath_.String())

}

