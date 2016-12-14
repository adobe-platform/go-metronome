package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"os"
	"strings"
	//"errors"
	"encoding/json"
	. "github.com/adobe-platform/go-metronome/metronome-cli/cli_support"
)

type CommandMap map[string]CommandParse

var commands CommandMap

// initialize the top level command map
func init() {
	commands = CommandMap{
		"job": CommandParse(new(JobTopLevel)),
		"run": CommandParse(new(RunsTopLevel)),
		"schedule": CommandParse(new(SchedTopLevel)),
		"metrics": CommandParse(new(Metrics)),
		"ping": CommandParse(new(Ping)),
	}
}

func Usage(msg string) {
	if msg != "" {
		logrus.Errorf(" %s ", msg)
	}
	options := []string{
		"job",
		"run",
		"schedule",
		"metrics",
		"ping",

	}
	fmt.Fprintf(os.Stderr, `USAGE

	 %s <global-options>  {%s|help} [<action options>|help]

COMMANDS:
	 `, os.Args[0], strings.Join(options, "|"))
	fmt.Fprintln(os.Stderr,"")
	for _, action := range options {
		commands[action].Usage(os.Stderr)
	}
	fmt.Fprintln(os.Stderr, `

GLOBAL OPTIONS:

		`)

	runtime := &Runtime{}
	runtime.Usage(os.Stderr)
	os.Exit(2)
}
func main() {
	logrus.SetOutput(os.Stderr)

	if len(os.Args) == 1 {
		Usage("")
	}
	keys := make([]string, 0, len(commands))
	for k := range commands {
		keys = append(keys, k)
	}
	index := -1
	var action string
	for v, value := range os.Args {
		if In(value, keys) {
			index = v
			action = value
			break
		}
	}
	runtime := &Runtime{}
	if index != -1 {
		var commonArgs []string
		if index > 1 {
			commonArgs = os.Args[1:index]
		} else {
			logrus.Debugf("No command args used\n")
		}
		logrus.Debugf("commonArgs %+v action: %s\n", commonArgs, action)
		if _, err := runtime.Parse(commonArgs); err != nil {
			Usage(err.Error())
		} else if action == "" {
			Usage("missing action")
		} else if commands[action] == nil {
			Usage(fmt.Sprintf("'%s' command not defined", action))
		}
		if runtime.Debug {
			logrus.SetLevel(logrus.DebugLevel)
		}
		var executorArgs []string
		if len(os.Args) > (index + 1) {
			executorArgs = os.Args[index + 1:]
		}
		logrus.Debugf("executorArgs %+v", executorArgs)
		if action == "help" {
			Usage("your help:")
		} else if executor, err := commands[action].Parse(executorArgs); err != nil {
			logrus.Fatalf("%s failed because %+v", action, err)
		} else {
			if result, err2 := executor.Execute(runtime); err2 != nil {
				logrus.Fatalf("action %s execution failed because %+v", action, err2)
			} else {
				logrus.Debugf("Result type: %T", result)

				switch result.(type){
				case json.RawMessage:
					var f interface{}
					by := result.(json.RawMessage)
					if err := json.Unmarshal(by, &f); err != nil {
						logrus.Infof(string(by))
					} else {
						if b2, err2 := json.MarshalIndent(f,"","  "); err2 != nil {
							logrus.Infof(string(by))
						} else {
							logrus.Infof(string(b2))
						}
					}
				default:
					if bb, err7 := json.MarshalIndent(result,"", "  "); err7 == nil {
						logrus.Infof("result %s\n", (string(bb)))
					}
				}
			}
		}
	} else {
		if len(os.Args) > 1 {
			Usage("You need to include a verb")
		} else {
			Usage("Nothing to do.  You need to choose an action\n")
		}
	}
}