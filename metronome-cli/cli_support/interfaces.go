package cli
import (
	"io"
)

// CommandExec is an interface returned by Parse when options are successfully parsed
// receiver Execute is passed the global options include the Metronome client interface
type CommandExec interface {
	Execute(runtime *Runtime) (interface{}, error)
}
// CommandParse
// implementor are passed arguments.
type CommandParse interface {
	Parse(args []string) (CommandExec, error)
	Usage(writer io.Writer)
}

