package cli

import (
	met "github.com/adobe-platform/go-metronome/metronome"
	"github.com/Sirupsen/logrus"
	"fmt"
	"strings"
	"errors"
	"net/url"
	"strconv"
)

//
// Job{Create|Update} take many parameters that must be validated and stored in nested structures
// These are set via flag.Var.  When using flag.Var, flag expects the passed pointer to implement the flag.Value interface
// So as to not effect the behavior of the actual types, these critical types are effectively aliased below to provide
// the correct command line handling for flag and the real type.  By doing so, it preserves the real types behavior
// flag.Var calls flag.Value interface of the provided interface{}
// The following light-weight types implement Value while preserving the Set/String symantics of the `real` type it alias.

// RunArgs - thin type providing Flags Value implementation for Metronome Run->Args
// type override to support parsing.  []string alias for met.Run.Args
// It implements flag.Value via Set/String
type RunArgs []string

// String - provide string helper
func (i *RunArgs) String() string {
	return fmt.Sprintf("%s", *i)
}

// Set - Value interface
func (i *RunArgs) Set(value string) error {
	logrus.Debugf("Args.Set %s", value)
	*i = append(*i, value)
	return nil
}
// type override to support parsing.  env alias' map[string]string
// It implements flag.Value via Set/String

// NvList - thin type providing Flags Value interface implementation for items needing map[string]string
type NvList map[string]string

// String - Value interface implementaion
func (list *NvList) String() string {
	return fmt.Sprintf("%s", *list)
}

// Set - Value interface implementation
func (list *NvList) Set(value string) error {
	logrus.Debugf("NvList %+v %s", list, value)
	nv := strings.Split(value, "=")
	if len(nv) != 2 {
		return errors.New("Environment vars should be NAME=VALUE")
	}
	logrus.Debugf("NvList %+v", nv)
	vv := (*list)
	vv[strings.TrimSpace(nv[0])] = strings.TrimSpace(nv[1])
	return nil
}

// ConstraintList - thin type providing Flags Value interface implementation for Metronome constraints
//   type override to support parsing.  ConstraintList alias' []met.Constraint
//   It implements flag.Value via Set/String
type ConstraintList [] met.Constraint

// String - Value interface implementation
func (list *ConstraintList) String() string {
	return fmt.Sprintf("%s", *list)
}

// Set - Value interface definition used with Flags
func (list *ConstraintList) Set(value string) error {
	con, err := met.StrToConstraint(value)
	if err != nil {
		return err
	}
	*list = ConstraintList(append(*list, *con))
	return nil

}


// VolumeList - thin type providing Flags Value interface implementation for Metronome volumes
type VolumeList [] met.Volume

// String - Value interface implementation
func (list *VolumeList) String() string {
	return fmt.Sprintf("%s", *list)
}

// Set - Value interface implementation
func (list *VolumeList) Set(value string) error {
	pieces := strings.Split(value, ":")
	if len(pieces) == 3 {

	} else if len(pieces) == 2 {
		pieces = append(pieces, "RO")
	}
	vol, err := met.NewVolume(pieces[0], pieces[1], pieces[2])
	if err != nil {
		return err
	}
	*list = append(*list, *vol)

	return nil
}

// ArtifactList - thin type providing Flags Value interface implementation for Metronome artifacts
type ArtifactList  []met.Artifact

// String - Value interface implementation
func (list *ArtifactList) String() string {
	return fmt.Sprintf("%s", *list)
}
// Set - Value interface implemention
func (list *ArtifactList) Set(value string) (err error) {
	var arty met.Artifact

	for _, pairs := range strings.Split(strings.TrimSpace(value), " ") {
		logrus.Debugf("pairs : %+v", pairs)
		kv := strings.SplitN(strings.TrimSpace(pairs), "=", 2)
		logrus.Debugf("kv=%+v", kv)
		switch strings.TrimSpace(kv[0]){
		case "url", "uri":
			ur, err := url.Parse(strings.TrimSpace(kv[1]));
			if err != nil {
				return err
			}
			arty.URI = ur.String()

		case "extract":
			if arty.Extract, err = strconv.ParseBool(kv[1]); err != nil {
				return err
			}
		case "executable":
			if arty.Executable, err = strconv.ParseBool(kv[1]); err != nil {
				return err
			}
		case "cache":
			if arty.Cache, err = strconv.ParseBool(kv[1]); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Unknown artifact '%s", kv[0])
		}

	}
	if arty.URI == "" {
		return errors.New("You must supply 'uri' for artifact")
	}
	*list = append(*list, arty)

	return nil
}