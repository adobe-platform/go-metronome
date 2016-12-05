#go-metronome
A go wrapper for metronome's API that includes a metronome-cli

# Status
The Metronome V1 interface is fully implemented.

# Getting start
```
go get github.com/adobe-platform/go-metronome
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

## Create a job
```
# ./cli job create -docker-image f4tq/dcos-tests:v0.31 -cmd '/usr/local/bin/dcos-tests --debug --term-wait 20 --http-addr :8095' -job-id "dcos.locust" --env "MON=test" --env "CONNECT=direct"

INFO[0000] result {"description":"","id":"dcos.locust","labels":{"location":"","owner":""},"run":{"cmd":"/usr/local/bin/dcos-tests --debug --term-wait 20 --http-addr :8095","cpus":0.2,"mem":128,"disk":128,"docker":{"image":"f4tq/dcos-tests:v0.31"},"env":{"CONNECT":"direct","MON":"test"},"maxLaunchDelay":900,"placement":{"constraints":[]},"restart":{"activeDeadlineSeconds":0,"policy":"NEVER"},"volumes":[]}}

```

## Update a job

Note> Job Update looks like Create.

```
# ./cli job update -docker-image f4tq/dcos-tests:v0.31 -cmd '/usr/local/bin/dcos-tests --debug --term-wait 20 --http-addr :8095' -job-id "dcos.locust" --env "MON=test" --env "CONNECT=direct"

INFO[0000] result {"id":"dcos.locust","description":"","labels":{},"run":{"cpus":0.2,"mem":128,"disk":128,"cmd":"/usr/local/bin/dcos-tests --debug --term-wait 20 --http-addr :8095","env":{"CONNECT":"direct","MON":"test4"},"placement":{"constraints":[]},"artifacts":[],"maxLaunchDelay":900,"docker":{"image":"f4tq/dcos-tests:v0.31"},"volumes":[],"restart":{"policy":"NEVER"}}}
```

## Get a defined job
```
# ./cli job get  -job-id "dcos.locust"

INFO[0000] result {"description":"","id":"dcos.locust","labels":{"location":"","owner":""},"run":{"cmd":"/usr/local/bin/dcos-tests --debug --term-wait 20 --http-addr :8095","cpus":0.2,"mem":128,"disk":128,"docker":{"image":"f4tq/dcos-tests:v0.31"},"env":{"CONNECT":"direct","MON":"test4"},"maxLaunchDelay":900,"placement":{"constraints":[]},"restart":{"activeDeadlineSeconds":0,"policy":"NEVER"},"volumes":[]}}
```
## Get all job definitions

Note:> Not to be confused with `running` jobs

```
# ./cli job ls
INFO[0000] result [{"description":"","id":"dcos.locust","labels":{"location":"","owner":""},"run":{"cmd":"/usr/local/bin/dcos-tests --debug --term-wait 20 --http-addr :8095","cpus":0.2,"mem":128,"disk":128,"docker":{"image":"f4tq/dcos-tests:v0.31"},"env":{"CONNECT":"direct","MON":"test4"},"maxLaunchDelay":900,"placement":{"constraints":[]},"restart":{"activeDeadlineSeconds":0,"policy":"NEVER"},"volumes":[]}}]

```

## Start a job **now**

```
# ./cli run start -job-id dcos.locust
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
# ./cli  run ls --job-id dcos.locust
INFO[0000] result [{"completedAt":null,"createdAt":"2016-12-05T18:27:45.287+0000","id":"20161205182745FuBxU","jobId":"dcos.locust","status":"ACTIVE","tasks":[{"id":"dcos_locust_20161205182745FuBxU.84419b42-bb18-11e6-a0f2-024273f73426","startedAt":"2016-12-05T18:27:46.399+0000","status":"TASK_RUNNING"}]}]
```


## A periodic job
This job is defined to echo a date

## Create the job definition
```
# ./cli job create -docker-image alpine:3.4 -cmd 'echo "testing $(date)' -job-id "foo.bar" -volume `cd test/;pwd`:/app
INFO[0000] result {"description":"","id":"foo.bar","labels":{"location":"","owner":""},"run":{"cmd":"echo \"testing $(date)\"","cpus":0.2,"mem":128,"disk":128,"docker":{"image":"alpine:3.4"},"maxLaunchDelay":900,"placement":{"constraints":[]},"restart":{"activeDeadlineSeconds":0,"policy":"NEVER"},"volumes":[{"containerPath":"/go/src/github.com/adobe-platform/go-metronome/cli/test","hostPath":"/app","mode":"RO"}]}}
```

### Schedule the periodic job definition to run every 2 minutes

Note> This doesn't start the job yet!

```
# ./cli schedule create -job-id foo.bar -sched-id ever2 -cron "*/2 * * * *" --start-deadline 60
INFO[0000] result {"id":"ever2","cron":"*/2 * * * *","concurrencyPolicy":"ALLOW","enabled":false,"startingDeadlineSeconds":60,"timezone":"Etc/GMT"}
```

### Update the schedule
Oops, I need the job to start with 45 seconds or we will wait until the next cycle

```
./cli schedule update -job-id foo.bar -sched-id ever2 -cron "*/2 * * * *" --start-deadline 45 --tz GMT
PUT {"id":"ever2","cron":"*/2 * * * *","concurrencyPolicy":"ALLOW","enabled":false,"startingDeadlineSeconds":45,"timezone":"GMT"}
INFO[0000] result {"id":"ever2","cron":"*/2 * * * *","concurrencyPolicy":"ALLOW","enabled":false,"startingDeadlineSeconds":45,"timezone":"GMT"}

```

### Delete the schedule
On 3rd thought, I don't like the schedule name so I'll delete the schedule
```
./cli schedule delete -job-id foo.bar --sched-id ever2
INFO[0000] result "OK"
```
Note:> You can't delete a schedule if the job is already running.  Stop it first.
 
### Run the job with the schedule
```
# ./cli job update -docker-image alpine:3.4 -cmd 'echo "testing $(date)"' -job-id "foo.bar" -volume `cd test/;pwd`:/app
INFO[0000] result {"id":"foo.bar","description":"","labels":{},"run":{"cpus":0.2,"mem":128,"disk":128,"cmd":"echo \"testing $(date)\"","env":{},"placement":{"constraints":[]},"artifacts":[],"maxLaunchDelay":900,"docker":{"image":"alpine:3.4"},"volumes":[{"containerPath":"/go/src/github.com/adobe-platform/go-metronome/cli/test","hostPath":"/app","mode":"RO"}],"restart":{"policy":"NEVER"}}}
```
