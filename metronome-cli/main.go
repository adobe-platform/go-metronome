package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"os"
	"strings"
	"errors"
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
	logrus.Errorf("usage: %s <global-options> <action: one of {%s}> [<action options>|help ]", os.Args[0], strings.Join([]string{
		"job",
		"run",
		"schedule",
		"metrics",
		"ping",
		"help",
	}, "|"))
	fmt.Println(" For more help, use ")
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
			panic(err)
		} else if action == "" {
			panic(errors.New("missing action"))
		} else if commands[action] == nil {
			panic(errors.New(fmt.Sprintf("'%s' command not defined", action)))
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
				if bb, err7 := json.Marshal(result); err7 == nil {
					logrus.Infof("result %s\n", (string(bb)))
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