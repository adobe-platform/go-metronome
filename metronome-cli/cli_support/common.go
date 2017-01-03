package cli
import (
	"flag"
	log "github.com/behance/go-logrus"
	"errors"
)

// Lightweight types for integrating with flags Value interface
// Support parsing re-use
// JobId, RunId, and SchedId are used in many calls.  Those 'types' are never used directly.  Instead they are part of other structs
//

// JobID - base type used to parse args for many commands requiring `job-id` parsing
type JobID string

// FlagSet - provide method for derived types to use common flags
func (id *JobID) FlagSet(flags *flag.FlagSet) *flag.FlagSet {
	flags.StringVar((*string)(id), "job-id", "", "Job Id")
	return flags
}
// Validate - validate state usually done as the last part of flag parsing
func (id *JobID) Validate() error {
	log.Debugf("JobId.Validate\n")
	if string(*id) == "" {
		return errors.New("job-id required")
	}
	return nil
}

// SchedID - lightweight type implementing FlagSet and Validate.  Common Metronome parameter
type SchedID string

// FlagSet - set up flag for setting schedule id
func (id *SchedID) FlagSet(flags *flag.FlagSet) *flag.FlagSet {
	flags.StringVar((*string)(id), "sched-id", "", "Schedule Id")
	return flags
}
// Validate - make sure that the flag was set
func (id *SchedID) Validate() error {
	if string(*id) == "" {
		return errors.New("sched-id required")
	}
	return nil
}
// RunID - is used in several REST calls
type RunID string

// FlagSet - Set the flag to collect run-id
func (id *RunID) FlagSet(flags *flag.FlagSet) *flag.FlagSet {
	flags.StringVar((*string)(id), "run-id", "", "Run Id")
	return flags
}
// Validate - that run-id has a non-zero length value
func (id *RunID) Validate() error {
	if string(*id) == "" {
		return errors.New("run-id required")
	}
	return nil
}

