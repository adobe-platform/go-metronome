package main

import (
	"encoding/json"
	//	"errors"
	"fmt"
	"log"
	//	"regexp"
	. "github.com/adobe-platform/go-metronome/metronome"
)

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
const data5 = `
{"description":"Example Application","id":"prod.example.app","labels":{"location":"olympus","owner":"zeus"},"run":{"artifacts":[{"uri":"http://foo.test.com/application.zip","extract":true,"executable":true,"cache":false}],"cmd":"nuke --dry --master local","args":["nuke","--dry","--master","local"],"cpus":1.5,"mem":32,"disk":128,"docker":{"image":"foo/bla:test"},"env":{"MON":"test","CONNECT":"direct"},"maxLaunchDelay":3600,"placement":{"constraints":[{"attribute":"rack","operator":"EQ","value":"rack-2"}]},"restart":{"activeDeadlineSeconds":120,"policy":"NEVER"},"user":"root","volumes":[{"containerPath":"/mnt/test","hostPath":"/etc/guest","mode":"RW"}]}}`

const data6_1 = `{"description":"Example Application","id":"prod.example.app","labels":{"location":"olympus","owner":"zeus"},"run":{"args":["nuke","--dry","--master","local"],"artifacts":[{"cache":false,"executable":true,"extract":true,"uri":"http://foo.test.com/application.zip"}],"cmd":"nuke --dry --master local","cpus":1.5,"disk":128,"docker":{"image":"foo/bla:test"},"env":{"CONNECT":"direct","MON":"test"},"maxLaunchDelay":3600,"mem":32,"placement":{"constraints":[]},"restart":{"activeDeadlineSeconds":120,"policy":"NEVER"},"user":"root","volumes":[{"containerPath":"/mnt/test","hostPath":"/etc/guest","mode":"RW"}]}}`
const data6_2 = `{"description":"Locust Test Application","id":"dcos.locust","labels":{"location":"olympus","owner":"zeus"},"run":{"artifacts":[],"cmd":"/usr/local/bin/dcos-tests --debug --term-wait 20 --http-addr :8095","cpus":0.5,"disk":128,"docker":{"image":"f4tq/dcos-tests:v0.31"},"env":{"CONNECT":"direct","MON":"test"},"maxLaunchDelay":3600,"mem":32,"placement":{"constraints":[]},"restart":{"activeDeadlineSeconds":120,"policy":"NEVER"},"user":"root","volumes":[]}}`

type j6 struct {
	v7 []int
}
//type foo interface {
//	Bar() []int
//}
func (self *j6) Bar() *[]int {
	return &self.v7
}
func main() {
	v := make([]Job, 0, 0)
	data6 := fmt.Sprintf("[%s,%s]", data6_1, data6_2)

	if err := json.Unmarshal([]byte(data6), &v); err != nil {
		log.Fatal("Unmarshal failed", err)
	} else {
		if res1B, err := json.Marshal(v[0]); err != nil {
			panic(err)
		} else {
			fmt.Printf(" >> 0) orig json: \n%s\n[0] final json: \n%s\n", data6_1, string(res1B))
		}

		if res1B, err := json.Marshal(v[1]); err != nil {
			panic(err)
		} else {
			fmt.Printf(" >> 1) orig json: \n%s\n[1]final json: \n%s\n", data6_2, string(res1B))
		}

	}
	main4()

}

func main4() {
	fmt.Println("test job reading from json, building up from code...")

	fmt.Printf("1. Test unmarshalling constraint from json:\n")
	a := &Constraint{}

	if err := json.Unmarshal([]byte(data), &a); err != nil {
		log.Fatal("Unmarshal failed", err)
	} else {
		if a.Attribute() == "jim gaffigan" && a.Operator() == EQ && a.Value() == "hot pockets" {
			log.Println("1. success")
		} else {
			log.Println("1. success")
		}
	}

	fmt.Printf("\n2. Read placement:\n  json: %s\n", data2)
	b := &Placement{}
	err2 := json.Unmarshal([]byte(data2), b)
	if err2 != nil {
		log.Fatal("Unmarshal failed", err2)
	}

	fmt.Printf("2. Resulting Constraint array %+v", b)
	fmt.Printf("3. Job <- json: %s\n", data5)
	var c Job
	err3 := json.Unmarshal([]byte(data5), &c)
	if err3 != nil {
		log.Fatal("Unmarshal failed", err3)
	}
	fmt.Printf("3. Golang job:\n %+v\n", c)
	if res1B, err := json.Marshal(c); err != nil {
		panic(err)
	} else {

		fmt.Printf("  3. Job -> json: %s\n", string(res1B))
	}

	fmt.Println("4.  Code minimal from scratch")

	if v, err5 := NewRun(0.4, 1, 1); err5 != nil {
		panic(err5)
	} else {
		fmt.Printf("  go rep: %+v\n", v)

		if js, err := json.Marshal(v); err != nil {
			panic(err)
		} else {
			fmt.Printf("  to json: %s\n", js)
		}
		args := v.Args()
		*args = append(*args, "foobar")

		v.AddArg("your momma")

		fmt.Printf("  Added args. %+v parent: %+v\n", args, v.Args())
		if js2, err := json.Marshal(v); err != nil {
			panic(err)
		} else {
			fmt.Printf("  4. to_json with args: %s\n\n", js2)
		}

	}
	//////
	fmt.Printf("\n5. Build Job to match 3) from golang alone:\n")
	if runnable, err5 := NewRun(1.5, 32, 128); err5 != nil {
		panic(err5)
	} else {

		runnable.SetDocker(&Docker{
			Image_: "foo/bla:test",
		}).SetEnv(map[string]string{
			"MON": "test",
			"CONNECT": "direct",
		}).SetArgs([]string{
			"nuke",
			"--dry",
			"--master",
			"local",
		}).SetCmd("nuke --dry --master local").SetPlacement(&Placement{
			Constraints_: []Constraint{
				Constraint{Attribute_: "rack", Operator_: EQ, Value_: "rack-2"} },

		}).SetArtifacts([] Artifact{
			Artifact{Uri_: "http://foo.test.com/application.zip", Extract_: true, Executable_ :true, Cache_: false},
		}).SetMaxLaunchDelay(
			3600,
		).SetRestart(&Restart{
			ActiveDeadlineSeconds_: 120, Policy_: "NEVER",

		}).SetVolumes([]Volume{
			Volume{Mode_:RW, HostPath_:"/etc/guest", ContainerPath_: "/mnt/test", },
		}).SetUser("root")

		if job, err6 := NewJob("prod.example.app", "Example Application", nil, runnable); err6 != nil {
			panic(err6)
		} else {
			job.SetLabel(Labels{
				Location: "olympus",
				Owner: "zeus",
			})
			if res1B, err := json.Marshal(job); err != nil {
				panic(err)
			} else {
				fmt.Printf(">> old\n<< new\n>>--\n%s\n<<--\n%s\n", data5, string(res1B))
			}
		}
	}
}

