package cli

import (
	met "github.com/adobe-platform/go-metronome/metronome"
	"github.com/Sirupsen/logrus"
	"fmt"
	"strings"
	"errors"
)

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