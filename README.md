#go-metronome
A go wrapper for metronome's API that includes a metronome-cli

# Status
The Metronome V1 interface is fully implemented.

# Getting started
You don't need golang installed if you have docker installed.

- If you want to install the cli locally
```
go get github.com/adobe-platform/go-metronome/metronome-cli
```
- If you want to install the API (sans cli) locally
```
go get github.com/adobe-platform/go-metronome/metronome
```

- If you want to build a docker image
```
# git clone https://github.com/adobe-platform/go-metronome.git
# cd go-metronome
# make build-container
```
  - Run the docker image just made
```
$(eval make run) job ls
```

- If you want a dev container (golang, deps)

```
make run-dev
```

- If you want to build a Linux and Darwin cli binary

```
make compile
```

> Yielding `go-metronome/metronome-cli-linux-amd64` and `go-metronome/metronome-cli-darwin-amd64`

# V1 Interface


```
type Metronome interface {
	// POST /v1/jobs
	CreateJob(*Job) (*Job, error)
	// DELETE /v1/jobs/$jobId
	DeleteJob(jobId string) (interface{}, error)
	// GET /v1/jobs/$jobId
	GetJob(jobId string) (*Job, error)
	// GET /v1/jobs
	Jobs() (*[]Job, error)
	// PUT /v1/jobs/$jobId
	JobUpdate(jobId string, job *Job) (interface{}, error)
	//
	// schedules
	// GET /v1/jobs/$jobId/runs
	RunLs(jobId string) (*[]JobStatus, error)
	// POST /v1/jobs/$jobId/runs
	RunStartJob(jobId string) (interface{}, error)
	// GET /v1/jobs/$jobId/runs/$runId
	RunStatusJob(jobId string, runId string) (*JobStatus, error)
	// POST /v1/jobs/$jobId/runs/$runId/action/stop
	RunStopJob(jobId string, runId string) (interface{}, error)

	//
	// Schedules
	//
	// POST /v1/jobs/$jobId/schedules
	JobScheduleCreate(jobId string, new *Schedule) (interface{}, error)
	// GET /v1/jobs/$jobId/schedules/$scheduleId
	JobScheduleGet(jobId string, schedId string) (*Schedule, error)
	// GET /v1/jobs/$jobId/schedules
	JobScheduleList(jobId string) (*[]Schedule, error)
	// DELETE /v1/jobs/$jobId/schedules/$scheduleId
	JobScheduleDelete(jobId string, schedId string) (interface{}, error)
	// PUT /v1/jobs/$jobId/schedules/$scheduleId
	JobScheduleUpdate(jobId string, schedId string, sched *Schedule) (interface{}, error)

	//  GET  /v1/metrics
	Metrics() (interface{}, error)
	//  GET /v1/ping
	Ping() (*string, error)
}
```

> Reference [Metronome V1 API](https://dcos.github.io/metronome/docs/generated/api.html#)

Simple client

```
   import met "github.com/adobe-platform/go-metronome/metronome"
   ...
   config := met.NewDefaultConfig()
   config.URL = "http://localhost:9000"
   if client, err := met.NewClient(config); err != nil {
      panic(err)			  
   } else {
      jobs := client.JobLs()
   }
```

# CLI
The following examples assume you've started metronome infrastructure as:
```
DOCKER_IP=10.0.2.15 docker-compose -f docker-compose.yml up
```

> DOCKER_IP should be the ip address of your `eth0` network interface

Using the docker-compose.yml file, you should see the following instances runnings.
```
f4c7f8ac85df        f4tq/metronome:0.9.1.4336d5b15cdd19-mesos-1.0.1   "/bin/sh -c '$APP_DIR"   13 hours ago        Up 13 hours                             mesoscompose_metronome_1
47548cc24717        mesosphere/mesos-master:1.0.1-2.0.93.ubuntu1404   "mesos-master --regis"   13 hours ago        Up 13 hours                             mesoscompose_master_1
a50c32104523        mesosphere/marathon:v1.3.0                        "./bin/start"            13 hours ago        Up 13 hours                             mesoscompose_marathon_1
9783a5e883f3        mesosphere/mesos-slave:1.0.1-2.0.93.ubuntu1404    "mesos-slave"            13 hours ago        Up 13 hours                             mesoscompose_slave-one_1
1bf0a4f476e0        bobrik/zookeeper                                  "/run.sh"                13 hours ago        Up 13 hours                             mesoscompose_zk_1
```


## Create a job
```
# metronome-cli/metronome-cli job create -docker-image f4tq/dcos-tests:v0.31 -cmd '/usr/local/bin/dcos-tests --debug --term-wait 20 --http-addr :8095' -job-id "dcos.locust" --env "MON=test" --env "CONNECT=direct"

INFO[0000] result {"description":"","id":"dcos.locust","labels":{"location":"","owner":""},"run":{"cmd":"/usr/local/bin/dcos-tests --debug --term-wait 20 --http-addr :8095","cpus":0.2,"mem":128,"disk":128,"docker":{"image":"f4tq/dcos-tests:v0.31"},"env":{"CONNECT":"direct","MON":"test"},"maxLaunchDelay":900,"placement":{"constraints":[]},"restart":{"activeDeadlineSeconds":0,"policy":"NEVER"},"volumes":[]}}

```

## Update a job

> Job Update looks like Create.

```
# metronome-cli/metronome-cli job update -docker-image f4tq/dcos-tests:v0.31 -cmd '/usr/local/bin/dcos-tests --debug --term-wait 20 --http-addr :8095' -job-id "dcos.locust" --env "MON=test" --env "CONNECT=direct"

INFO[0000] result {"id":"dcos.locust","description":"","labels":{},"run":{"cpus":0.2,"mem":128,"disk":128,"cmd":"/usr/local/bin/dcos-tests --debug --term-wait 20 --http-addr :8095","env":{"CONNECT":"direct","MON":"test4"},"placement":{"constraints":[]},"artifacts":[],"maxLaunchDelay":900,"docker":{"image":"f4tq/dcos-tests:v0.31"},"volumes":[],"restart":{"policy":"NEVER"}}}
```

## Get a defined job
```
# metronome-cli/metronome-cli job get  -job-id "dcos.locust"

INFO[0000] result {"description":"","id":"dcos.locust","labels":{"location":"","owner":""},"run":{"cmd":"/usr/local/bin/dcos-tests --debug --term-wait 20 --http-addr :8095","cpus":0.2,"mem":128,"disk":128,"docker":{"image":"f4tq/dcos-tests:v0.31"},"env":{"CONNECT":"direct","MON":"test4"},"maxLaunchDelay":900,"placement":{"constraints":[]},"restart":{"activeDeadlineSeconds":0,"policy":"NEVER"},"volumes":[]}}
```
## Get all job definitions

> Not to be confused with `running` jobs

```
# metronome-cli/metronome-cli job ls
INFO[0000] result [{"description":"","id":"dcos.locust","labels":{"location":"","owner":""},"run":{"cmd":"/usr/local/bin/dcos-tests --debug --term-wait 20 --http-addr :8095","cpus":0.2,"mem":128,"disk":128,"docker":{"image":"f4tq/dcos-tests:v0.31"},"env":{"CONNECT":"direct","MON":"test4"},"maxLaunchDelay":900,"placement":{"constraints":[]},"restart":{"activeDeadlineSeconds":0,"policy":"NEVER"},"volumes":[]}}]

```

## Start a job **now**

```
# metronome-cli/metronome-cli run start -job-id dcos.locust
INFO[0000] result {"completedAt":null,"createdAt":"2016-12-05T19:33:20.586+0000","id":"20161205193320NR9q1","jobId":"dcos.locust","status":"INITIAL","tasks":[]}

```


###  docker-compose users

If you're using docker-compose.yml to present the metronome/mesos environment, then `docker ps` will show your running container.  

```
# docker ps

42202f7cba7a        f4tq/dcos-tests:v0.31                             "/bin/sh -c '/usr/loc"   22 minutes ago      Up 22 minutes                           mesos-22357193-cfdd-48e6-b3ed-288c2a906bb0-S0.66f17489-615b-4b52-8a8c-395b3c6ccaa2
f4c7f8ac85df        f4tq/metronome:0.9.1.4336d5b15cdd19-mesos-1.0.1   "/bin/sh -c '$APP_DIR"   3 hours ago         Up 3 hours                              mesoscompose_metronome_1
47548cc24717        mesosphere/mesos-master:1.0.1-2.0.93.ubuntu1404   "mesos-master --regis"   3 hours ago         Up 3 hours                              mesoscompose_master_1
a50c32104523        mesosphere/marathon:v1.3.0                        "./bin/start"            3 hours ago         Up 3 hours                              mesoscompose_marathon_1
9783a5e883f3        mesosphere/mesos-slave:1.0.1-2.0.93.ubuntu1404    "mesos-slave"            3 hours ago         Up 3 hours                              mesoscompose_slave-one_1
1bf0a4f476e0        bobrik/zookeeper                                  "/run.sh"                3 hours ago         Up 3 hours                              mesoscompose_zk_1

```

> The new, metronome created job is `mesos-22357193-cfdd-48e6-b3ed-288c2a906bb0-S0.66f17489-615b-4b52-8a8c-395b3c6ccaa2`

`f4tq/dcos-test:v0.3` listens on port 8095 as we defined in the job definitions --cmd option

```
http GET http://localhost:8095/sleep/1
HTTP/1.1 200 OK
Connection: keep-alive
Content-Length: 110
Content-Type: text/plain
Date: Mon, 05 Dec 2016 19:57:52 GMT
Keep-Alive: timeout=60
Server: gophr

[2016-12-05 19:57:52.254281597 +0000 UTC] slept for 1 seconds starting @2016-12-05 19:57:51.24923488 +0000 UTC
```


## Status a job *run*


```
# metronome-cli/metronome-cli  run ls --job-id dcos.locust
INFO[0000] result [{"completedAt":null,"createdAt":"2016-12-05T18:27:45.287+0000","id":"20161205182745FuBxU","jobId":"dcos.locust","status":"ACTIVE","tasks":[{"id":"dcos_locust_20161205182745FuBxU.84419b42-bb18-11e6-a0f2-024273f73426","startedAt":"2016-12-05T18:27:46.399+0000","status":"TASK_RUNNING"}]}]
```


## A periodic job
This job is defined to echo a date

## Create the job definition
```
# metronome-cli/metronome-cli job create -docker-image alpine:3.4 -cmd 'echo "testing $(date)' -job-id "foo.bar" -volume `cd test/;pwd`:/app
INFO[0000] result {"description":"","id":"foo.bar","labels":{"location":"","owner":""},"run":{"cmd":"echo \"testing $(date)\"","cpus":0.2,"mem":128,"disk":128,"docker":{"image":"alpine:3.4"},"maxLaunchDelay":900,"placement":{"constraints":[]},"restart":{"activeDeadlineSeconds":0,"policy":"NEVER"},"volumes":[{"containerPath":"/go/src/github.com/adobe-platform/go-metronome/cli/test","hostPath":"/app","mode":"RO"}]}}
```

### Schedule the periodic job definition to run every 2 minutes

> Note: This doesn't start the job yet!

```
# metronome-cli/metronome-cli schedule create -job-id foo.bar -sched-id ever2 -cron "*/2 * * * *" --start-deadline 60
INFO[0000] result {"id":"ever2","cron":"*/2 * * * *","concurrencyPolicy":"ALLOW","enabled":false,"startingDeadlineSeconds":60,"timezone":"Etc/GMT"}
```

### Update the schedule
Oops, I need the job to start with 45 seconds or we will wait until the next cycle

```
metronome-cli/metronome-cli schedule update -job-id foo.bar -sched-id ever2 -cron "*/2 * * * *" --start-deadline 45 --tz GMT
PUT {"id":"ever2","cron":"*/2 * * * *","concurrencyPolicy":"ALLOW","enabled":false,"startingDeadlineSeconds":45,"timezone":"GMT"}
INFO[0000] result {"id":"ever2","cron":"*/2 * * * *","concurrencyPolicy":"ALLOW","enabled":false,"startingDeadlineSeconds":45,"timezone":"GMT"}

```

### Delete the schedule
On 3rd thought, I don't like the schedule name so I'll delete the schedule
```
metronome-cli/metronome-cli schedule delete -job-id foo.bar --sched-id ever2
INFO[0000] result "OK"
```
> You can't delete a schedule if the job is already running.  Stop it first.
 
### Run the job with the schedule
```
# metronome-cli/metronome-cli job update -docker-image alpine:3.4 -cmd 'echo "testing $(date)"' -job-id "foo.bar" -volume `cd test/;pwd`:/app
INFO[0000] result {"id":"foo.bar","description":"","labels":{},"run":{"cpus":0.2,"mem":128,"disk":128,"cmd":"echo \"testing $(date)\"","env":{},"placement":{"constraints":[]},"artifacts":[],"maxLaunchDelay":900,"docker":{"image":"alpine:3.4"},"volumes":[{"containerPath":"/go/src/github.com/adobe-platform/go-metronome/cli/test","hostPath":"/app","mode":"RO"}],"restart":{"policy":"NEVER"}}}
```

# Using with dc/os
This guide assumes you work at Adobe and you need to access a bastion host to reach via your dcos cluster.  It also assumes that you are accessing the DC/OS universe, mesos master, marathon and metronome via the tunnel.

## Set up an ssh tunnel
- Get ssh out permission from Juniper

- ssh to you bastion like

```
ssh -o DynamicForward=localhost:1200 -N ec2-user@$your_bastion_host &
```

- Set up an http proxy that knows how 
  - Install polipo

``` 
  vagrant@osx $ apt-get install polipo
```
  - Run the proxy so that all requests use the ssh-base SOCKS proxy
```
sudo polipo socksParentProxy=localhost:1200 diskCacheRoot=/dev/null
```

> By default, `polipo` listens on port 8123 for proxy connections

- Install dcos cli

Left to the reader



- Make dc/os cli use the proxy and login.

For example:

```
# http_proxy=localhost:8123  dcos auth login


Please go to the following link in your browser:

    http://internal-f4tq2-san-Internal-151LVCIS6CMZZ-510625652.us-east-1.elb.amazonaws.com/login?redirect_uri=urn:ietf:wg:oauth:2.0:oob

Enter authentication token: eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6Ik9UQkVOakZFTWtWQ09VRTRPRVpGTlRNMFJrWXlRa015Tnprd1JrSkVRemRCTWpBM1FqYzVOZyJ9.eyJlbWFpbCI6ImY0dHFAeWFob28uY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImlzcyI6Imh0dHBzOi8vZGNvcy5hdXRoMC5jb20vIiwic3ViIjoiZ2l0aHVifDgyNjEzMCIsImF1ZCI6IjN5RjVUT1N6ZGxJNDVRMXhzcHh6ZW9HQmU5Zk54bTltIiwiZXhwIjoxNDgxNDMwNjU2LCJpYXQiOjE0ODA5OTg2NTZ9.R6NvhVAYfSEl2MB0yhtSn-j9Lf8eMMrAxZrJ0dN-QoR3qRwoRKVyBAWRSLej5iY7e-hN3t3v5Ra24vteouowvbcDVNbDLCoXLxRCbLSHoHeIietl3bK8xZRlwRYCBcPJmdp3uQ46gZVayL4jTZ8DhdeRYkvhmG-XG4eaSB9WtUQUgiGO4lHFvnF0J8979_4-LTtB-EVu6FED6aOG0D5koj4wb7zgNV2orQlzAjY5u5wlIMHQ626aFI1nWoeUb1GTZAxe6GCSGtJCa1HfHIUD0XGV6QmculdwM4i6wwRZDlP9N1UJnIPZPZLa5SffrbaZsrov3Qdms_I04rngYVhs_A
Login successful!

```

>  Use firefox to get your token - which must also use the SOCKS 5 proxy.   Launch `/Applications/Firefox.app/Contents/MacOS/firefox-bin -profilemanager`. Navigate to Preferences->Advanced->Network->Settings. Click `Manual proxy`; SOCKS Host: `localhost:1200`; Click `SOCKS v5`; click `Remote DNS`


- Now the token with metronome-cli 

  - To ping
```
HTTP_PROXY=localhost:8123  ./metronome-cli-linux-amd64 --debug --metronome-url "$( dcos config show core.dcos_url)/service/metronome" --authorization "$( dcos config show core.dcos_acs_token)" ping

INFO[0000] result "pong"
```
  - To get metrics
```
HTTP_PROXY=localhost:8123 ~/go/src/github.com/adobe-platform/go-metronome/metronome-cli-linux-amd64  --metronome-url "$( dcos config show core.dcos_url)/service/metronome" --authorization "$( dcos config show core.dcos_acs_token)" metrics
INFO[0000] result {"version":"3.0.0","gauges":{"jvm.buffers.direct.capacity":{"value":185715},"jvm.buffers.direct.count":{"value":15},"jvm.buffers.direct.used":{"value":185716},"jvm.buffers.mapped.capacity":{"value":0},"jvm.buffers.mapped.count":{"value":0},"jvm.buffers.mapped.used":{"value":0},"jvm.gc.PS-MarkSweep.count":{"value":3},"jvm.gc.PS-MarkSweep.time":{"value":427},"jvm.gc.PS-Scavenge.count":{"value":475},"jvm.gc.PS-Scavenge.time":{"value":1189},"jvm.memory.heap.committed":{"value":174063616},"jvm.memory.heap.init":{"value":123731968},"jvm.memory.heap.max":{"value":1744830464},"jvm.memory.heap.usage":{"value":0.032752701869377725},"jvm.memory.heap.used":{"value":57147912},"jvm.memory.non-heap.committed":{"value":101007360},"jvm.memory.non-heap.init":{"value":2555904},"jvm.memory.non-heap.max":{"value":-1},"jvm.memory.non-heap.usage":{"value":-9.9773352E7},"jvm.memory.non-heap.used":{"value":99773352},"jvm.memory.pools.Code-Cache.committed":{"value":22282240},"jvm.memory.pools.Code-Cache.init":{"value":2555904},"jvm.memory.pools.Code-Cache.max":{"value":251658240},"jvm.memory.pools.Code-Cache.usage":{"value":0.08743769327799479},"jvm.memory.pools.Code-Cache.used":{"value":22004416},"jvm.memory.pools.Compressed-Class-Space.committed":{"value":9871360},
--snip -- 
```
  - To get job list

```
HTTP_PROXY=localhost:8123 ~/go/src/github.com/adobe-platform/go-metronome/metronome-cli-linux-amd64  --metronome-url "$( dcos config show core.dcos_url)/service/metronome" --authorization "$( dcos config show core.dcos_acs_token)" job ls


INFO[0000] result [{"description":"Installs VAMP and dependencies","id":"install-vamp","labels":{"location":"","owner":""},"run":{"artifacts":[{"uri":"https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64","executable":true,"extract":false,"cache":true},{"uri":"https://gist.githubusercontent.com/mhausenblas/bb967625088902874d631eaa502573cb/raw/4829525ab7700645166f7c47843cb351e3d2a807/install-vamp-09.sh","executable":true,"extract":false,"cache":false},{"uri":"https://gist.githubusercontent.com/mhausenblas/bb967625088902874d631eaa502573cb/raw/7e738db72693716a246c29abd320d67c5a4ec74b/vamp09-es.json","executable":false,"extract":false,"cache":false},{"uri":"https://gist.githubusercontent.com/mhausenblas/bb967625088902874d631eaa502573cb/raw/7e738db72693716a246c29abd320d67c5a4ec74b/vamp09.json","executable":true,"extract":false,"cache":true},{"uri":"https://gist.githubusercontent.com/mhausenblas/bb967625088902874d631eaa502573cb/raw/7e738db72693716a246c29abd320d67c5a4ec74b/vamp09-gateway.json","executable":false,"extract":false,"cache":false}],"cmd":"mv jq-linux64 jq \u0026\u0026 ./install-vamp-09.sh","cpus":0.5,"mem":100,"disk":0,"maxLaunchDelay":3600,"placement":{"constraints":[]},"restart":{"activeDeadlineSeconds":0,"policy":"NEVER"},"volumes":[]}}]
```


# Detailed CLI Usage

## No options
```
/usr/local/bin/metronome-cli-linux-amd64 <global-options> <action: one of {job|run|schedule|metrics|ping|help}> [<action options>|help ] 
 For more help, use 
  -debug
        Turn on debug
  -metronome-url string
        Set the Metronome address (default "http://localhost:9000")
```
## job sub menu
```
docker run -i --rm --net host -t adobe-platform/go-metronome:029ef7418691d812dacc4a01ded16598cfbe7ccb /usr/local/bin/metronome-cli-linux-amd64 job       
FATA[0000] job failed because job subcommand required

job  usage:
job {create|delete|update|ls|get|schedules|schedule|help}

```

###  `job create`

```
docker run -i --rm --net host -t adobe-platform/go-metronome:029ef7418691d812dacc4a01ded16598cfbe7ccb /usr/local/bin/metronome-cli-linux-amd64 job create

job create usage:
  -arg value
        Adds Arg metrononome->Job->Run->Args. You can call more than once
  -cmd string
        Command to run
  -constraint value
        Add Constraint used to construct Job->Run->[]Constraint
  -cpus float
        cpus (default 0.2)
  -description string
        Job Description - optional
  -disk int
        disk (default 128)
  -docker-image string
        Docker Image (default "alpine:3.4")
  -env value
        VAR=VAL . Adds Volume passed to metrononome->Job->Run->Volumes.  You can call more than once
  -job-id string
        Job Id
  -label value
        Location=xxx; Owner=yyy
  -max-launch-delay int
        Max Launch delay.  minimum 1 (default 900)
  -memory int
        memory (default 128)
  -restart-active-deadline-seconds int
        If the job fails, how long should we try to restart the job. If no value is set, this means forever.
  -restart-policy string
        Restart policy on job failure: NEVER or ALWAYS (default "NEVER")
  -run-now
        Run this job now, otherwise it is created as unscheduled
  -user string
        user to run as (default "root")
  -volume value
        /host:/container:{RO|RW} . Adds Volume passed to metrononome->Job->Run->Volumes. You can call more than once
```

## schedule submenu
```
docker run -i --rm --net host -t adobe-platform/go-metronome:029ef7418691d812dacc4a01ded16598cfbe7ccb /usr/local/bin/metronome-cli-linux-amd64 schedule 
FATA[0000] schedule failed because sub command required

schedule  usage:
schedule {create|delete|update|get|ls}  

          create  <options>
          delete  <options>
          update  <options>
          get     <options>
          ls


```

### schedule create
```
docker run -i --rm --net host -t adobe-platform/go-metronome:029ef7418691d812dacc4a01ded16598cfbe7ccb /usr/local/bin/metronome-cli-linux-amd64 schedule create
FATA[0000] schedule failed because Missing JobId in JobScheduleCreate


schedule create usage:
  -concurrency-policy string
        Schedule concurrency.  One of ALLOW,FORBID,REPLACE (default "ALLOW")
  -cron string
        Schedule Cron
  -enabled
        Enable the schedule (default true)
  -job-id string
        Job Id
  -sched-id string
        Schedule Id
  -start-deadline int
        Schedule deadline
  -tz string
        Schedule time zone (default "GMT")

```

## run submenu
```
docker run -i --rm --net host -t adobe-platform/go-metronome:029ef7418691d812dacc4a01ded16598cfbe7ccb /usr/local/bin/metronome-cli-linux-amd64 run            
FATA[0000] run failed because sub command required

  usage:
run <action> [options]:

          start [options]
          stop  [options]
          ls
          get [options]

          Call run <action> help for more on a sub-command
```
### run create
```
docker run -i --rm --net host -t adobe-platform/go-metronome:029ef7418691d812dacc4a01ded16598cfbe7ccb /usr/local/bin/metronome-cli-linux-amd64 run start 
FATA[0000] run failed because job-id required


 start usage:
  -job-id string
        Job Id
```
