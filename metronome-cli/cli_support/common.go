package cli
import (
	"flag"
	"github.com/Sirupsen/logrus"
	"errors"
)

//
// Support parsing re-use
// JobId, RunId, and SchedId are used in many calls.  Those 'types' are never used directly.  Instead they are part of other structs
//
// base class used to parse args for many commands requiring `job-id` parsing
type JobId string

func (self *JobId) FlagSet(flags *flag.FlagSet) *flag.FlagSet {
	flags.StringVar((*string)(self), "job-id", "", "Job Id")
	return flags
}
func (self *JobId) Validate() error {
	logrus.Debugf("JobId.Validate\n")
	if string(*self) == "" {
		return errors.New("job-id required")
	}
	return nil
}

type SchedId string

func (self *SchedId) FlagSet(flags *flag.FlagSet) *flag.FlagSet {
	flags.StringVar((*string)(self), "sched-id", "", "Schedule Id")
	return flags
}
func (self *SchedId) Validate() error {
	if string(*self) == "" {
		return errors.New("sched-id required")
	}
	return nil
}
// RunId is used in several REST calls
type RunId string

func (self *RunId) FlagSet(flags *flag.FlagSet) *flag.FlagSet {
	flags.StringVar((*string)(self), "run-id", "", "Run Id")
	return flags
}
func (self *RunId) Validate() error {
	if string(*self) == "" {
		return errors.New("run-id required")
	}
	return nil
}

