#go-metronome
A go wrapper for metronome's API that includes a metronome-cli

# Status
The Metronome V1 interface is fully implemented.

# Getting started
You don't need golang installed if you have docker installed.

- If you want to install locally
```
go get github.com/adobe-platform/go-metronome
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

Note:> DOCKER_IP should be the ip address of your `eth0` network interface

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

Note> Job Update looks like Create.

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

Note:> Not to be confused with `running` jobs

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

Note> The new job is `mesos-22357193-cfdd-48e6-b3ed-288c2a906bb0-S0.66f17489-615b-4b52-8a8c-395b3c6ccaa2`

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

Note> This doesn't start the job yet!

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
Note:> You can't delete a schedule if the job is already running.  Stop it first.
 
### Run the job with the schedule
```
# metronome-cli/metronome-cli job update -docker-image alpine:3.4 -cmd 'echo "testing $(date)"' -job-id "foo.bar" -volume `cd test/;pwd`:/app
INFO[0000] result {"id":"foo.bar","description":"","labels":{},"run":{"cpus":0.2,"mem":128,"disk":128,"cmd":"echo \"testing $(date)\"","env":{},"placement":{"constraints":[]},"artifacts":[],"maxLaunchDelay":900,"docker":{"image":"alpine:3.4"},"volumes":[{"containerPath":"/go/src/github.com/adobe-platform/go-metronome/cli/test","hostPath":"/app","mode":"RO"}],"restart":{"policy":"NEVER"}}}
```

# Using with dc/os
This guide assumes you work at Adobe and you need to access a bastion host to reach via a bastion host.  It also assumes that you are accessing the DC/OS universe, mesos master, marathon and metronome via the tunnel.

## Set up an ssh tunnel
- Get permission from Juniper
- ssh to you bastion like
```
ssh -o DynamicForward=localhost:1200 -N &
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

Note:> By default, `polipo` listens on port 8123 for proxy connections

- Install dcos cli

Left to the reader



- Make dc/os cli use the proxy and login.
```
http_proxy=localhost:8123  bin/exec dcos auth login

Please go to the following link in your browser:

    http://internal-f4tq2-san-Internal-151LVCIS6CMZZ-510625652.us-east-1.elb.amazonaws.com/login?redirect_uri=urn:ietf:wg:oauth:2.0:oob

Enter authentication token: eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6Ik9UQkVOakZFTWtWQ09VRTRPRVpGTlRNMFJrWXlRa015Tnprd1JrSkVRemRCTWpBM1FqYzVOZyJ9.eyJlbWFpbCI6ImY0dHFAeWFob28uY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImlzcyI6Imh0dHBzOi8vZGNvcy5hdXRoMC5jb20vIiwic3ViIjoiZ2l0aHVifDgyNjEzMCIsImF1ZCI6IjN5RjVUT1N6ZGxJNDVRMXhzcHh6ZW9HQmU5Zk54bTltIiwiZXhwIjoxNDgxNDMwNjU2LCJpYXQiOjE0ODA5OTg2NTZ9.R6NvhVAYfSEl2MB0yhtSn-j9Lf8eMMrAxZrJ0dN-QoR3qRwoRKVyBAWRSLej5iY7e-hN3t3v5Ra24vteouowvbcDVNbDLCoXLxRCbLSHoHeIietl3bK8xZRlwRYCBcPJmdp3uQ46gZVayL4jTZ8DhdeRYkvhmG-XG4eaSB9WtUQUgiGO4lHFvnF0J8979_4-LTtB-EVu6FED6aOG0D5koj4wb7zgNV2orQlzAjY5u5wlIMHQ626aFI1nWoeUb1GTZAxe6GCSGtJCa1HfHIUD0XGV6QmculdwM4i6wwRZDlP9N1UJnIPZPZLa5SffrbaZsrov3Qdms_I04rngYVhs_A
Login successful!

```

Note:>  Use firefox to get your token - which must also use the SOCKS 5 proxy.   `/Applications/Firefox.app/Contents/MacOS/firefox-bin -profilemanager` 


- Use the token to get `/v1/jobs` from the DC/OS' metronome
```
http_proxy=localhost:8123 http GET $(bin/exec dcos config show core.dcos_url)/service/metronome/v1/jobs Authorization:token=$(bin/exec dcos config show core.dcos_acs_token)

HTTP/1.1 200 OK
Connection: keep-alive
Content-Length: 1210
Content-Type: application/json
Date: Tue, 06 Dec 2016 04:42:13 GMT
Server: openresty/1.9.15.1

[
    {
        "description": "Installs VAMP and dependencies", 
        "id": "install-vamp", 
        "labels": {}, 
        "run": {
            "artifacts": [
                {
                    "cache": true, 
                    "executable": true, 
                    "extract": false, 
                    "uri": "https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64"
                }, 
-- snip --
```

- Now use `metronome-cli`

http_proxy=localhost:8123 