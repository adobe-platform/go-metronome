package metronome_test

import (
	"net/http"
	//"time"

	. "github.com/adobe-platform/go-metronome/metronome"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/ghttp"
	//"fmt"
	"bytes"
	"encoding/json"
	"fmt"
//	"net/url"
//	"github.com/onsi/gomega/types"
	"strconv"
)


var _ = Describe("Jobs", func() {
	var (
		config_stub Config
		client Metronome
		server      *ghttp.Server
		sched Schedule
		status JobStatus
	)
	// make a Schedule
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, `{"id":"every2","cron":"*/2 * * * *","concurrencyPolicy":"ALLOW","enabled":true,"startingDeadlineSeconds":60,"timezone":"Etc/GMT"}`)
	json.Unmarshal(buf.Bytes(), &sched)

	// make a JobStatus
	buf = new(bytes.Buffer)
	fmt.Fprintf(buf, `{
  "completedAt": null,
  "createdAt": "2016-07-15T13:02:59.735+0000",
  "id": "20160715130259A34HX",
  "jobId": "prod",
  "status": "STARTING",
  "tasks": []
}`)
	json.Unmarshal(buf.Bytes(), &status)
	// make an array of Jobs
	var allJobs []*Job
	buf = new(bytes.Buffer)
	fmt.Fprint(buf, `[
					    {
						"description": "Job with arguments",
						"id": "job.with.arguments",
						"labels": {
						    "Location": "olympus",
						    "Owner": "zeus"
						},
						"run": {
						    "args": [
							"nuke",
							"--dry",
							"--master",
							"local"
						    ],
						    "artifacts": [
							{
							    "cache": false,
							    "executable": true,
							    "extract": true,
							    "uri": "http://foo.test.com/application.zip"
							}
						    ],
						    "cmd": "nuke --dry --master local",
						    "cpus": 1.5,
						    "disk": 32,
						    "docker": {
							"image": "foo/bla:test"
						    },
						    "env": {
							"CONNECT": "direct",
							"MON": "test"
						    },
						    "maxLaunchDelay": 3600,
						    "mem": 128,
						    "placement": {
                        				"constraints": [
                            					{"attribute": "rack", "operator": "EQ", "value": "rack-2"}
                        				]
                    				    },
						    "restart": {
							"activeDeadlineSeconds": 120,
							"policy": "NEVER"
						    },
						    "user": "root",
						    "volumes": [
							{
							    "containerPath": "/mnt/test",
							    "hostPath": "/etc/guest",
							    "mode": "RW"
							}
						    ]
						}
					    },
					    {
						"description": "Job without arguments",
						"id": "job.without.arguments",
						"labels": {
						    "Location": "olympus",
						    "Owner": "zeus"
						},
						"run": {

						    "cmd": "/usr/local/bin/dcos-tests --debug --term-wait 20 --http-addr :8095",
						    "cpus": 0.5,
						    "disk": 128,
						    "docker": {
							"image": "f4tq/dcos-tests:v0.31"
						    },
						    "env": {
							"CONNECT": "direct",
							"MON": "test"
						    },
						    "maxLaunchDelay": 3600,
						    "mem": 32,
						    "restart": {
							"activeDeadlineSeconds": 120,
							"policy": "NEVER"
						    },
						    "user": "root"
						}
					    }
					]`)
	json.Unmarshal(buf.Bytes(), &allJobs)
	BeforeEach(func() {
		server = ghttp.NewServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/v1/jobs"),
			),
		)

		config_stub = Config{
			URL:            server.URL(),
			Debug:          false,
			RequestTimeout: 5,
		}

		// This will make a request and I dont know how to reset it
		// All checks for number of requests need to add one
		client, _ = NewClient(config_stub)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Jobs", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v1/jobs"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, allJobs),
				),
			)
		})

		It("Makes a request to get all jobs", func() {
			client.Jobs()
			Expect(server.ReceivedRequests()).To(HaveLen(2))
		})

		It("Correctly unmarshalls the response", func() {

			jobs, _ := client.Jobs()
			//var jj Job = (*jobs)[0]
			Expect((*jobs)[0]).To(Equal(Job{ID_: "job.with.arguments",
				Description_: "Job with arguments",
				Labels_: &Labels{
					"Location": "olympus",
					"Owner": "zeus",
				},
				Run_: &Run{
					Artifacts_: []Artifact{
						Artifact{Uri_: "http://foo.test.com/application.zip", Extract_: true, Executable_ :true, Cache_: false},
					},
					Cmd_: "nuke --dry --master local",
					Args_:[]string{
						"nuke",
						"--dry",
						"--master",
						"local",
					},
					Cpus_: 1.5,
					Mem_: 128,
					Disk_: 32,
					Docker_ : &Docker{
						Image_: "foo/bla:test",
					},
					Env_:  map[string]string{
						"MON": "test",
						"CONNECT": "direct",
					},
					MaxLaunchDelay_: 3600,
					Placement_: &Placement{
						Constraints_: []Constraint{
							Constraint{Attribute_: "rack", Operator_: EQ, Value_: "rack-2"} },

					},
					Restart_: &Restart{
						ActiveDeadlineSeconds_: 120, Policy_: "NEVER",

					},
					User_: "root",
					Volumes_: []Volume{
						Volume{Mode_:RW, HostPath_:"/etc/guest", ContainerPath_: "/mnt/test" },
					},
				},
			}))

			Expect((*jobs)[1]).To(Equal(
				Job{Description_: "Job without arguments",
					ID_: "job.without.arguments",
					Labels_: &Labels{
						"Location": "olympus",
						"Owner": "zeus",
					},
					Run_: &Run{Cmd_: "/usr/local/bin/dcos-tests --debug --term-wait 20 --http-addr :8095",
						Cpus_: 0.5,
						Disk_: 128,
						Docker_: &Docker{
							Image_: "f4tq/dcos-tests:v0.31",
						},
						Env_:   map[string]string{
							"MON": "test",
							"CONNECT": "direct",
						},
						MaxLaunchDelay_: 3600,
						Mem_: 32,
						Restart_: &Restart{
							ActiveDeadlineSeconds_: 120,
							Policy_: "NEVER",
						},
						User_: "root",
					},

				}))

		})
	})
	Describe("HiddenAPI", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					// , "_timestamp=1481414058857&embed=history&embed=historySummary"
					ghttp.VerifyRequest("GET", "/v1/jobs/foo.bar"),
					func(w http.ResponseWriter, req *http.Request) {
						values := req.URL.Query()
						Expect(values["_timestamp"]).NotTo(Equal(nil))
						Expect(values["_timestamp"]).To(HaveLen(1))
						tm, err := strconv.Atoi(values["_timestamp"][0])
						Expect(err).ShouldNot(HaveOccurred(),"Expected _timestamp >= 1481414058857")
						Expect(tm > 1481414058857 ).To(BeTrue())

					},
					ghttp.RespondWith(http.StatusOK, `
{"id":"foo.bar","description":"","labels":{},"run":{"cpus":0.2,"mem":128,"disk":128,"cmd":"echo \"testing $(date)\"","env":{},"placement":{"constraints":[]},"artifacts":[],"maxLaunchDelay":900,"docker":{"image":"alpine:3.4"},"volumes":[{"containerPath":"/var/lib/tmp","hostPath":"/app","mode":"RO"}],"restart":{"policy":"NEVER","activeDeadlineSeconds":0}},"schedules":[{"id":"every2","cron":"*/2 * * * *","timezone":"Etc/GMT","startingDeadlineSeconds":60,"concurrencyPolicy":"ALLOW","enabled":true,"nextRunAt":"2016-12-12T20:16:00.000+0000"}],"activeRuns":[{"id":"20161212192759dliHA","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:27:59.057+0000","completedAt":null,"tasks":[]},{"id":"20161212192200ouvXM","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:22:00.009+0000","completedAt":null,"tasks":[]},{"id":"20161212200759oAg7W","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T20:07:59.397+0000","completedAt":null,"tasks":[]},{"id":"20161212194759esoLp","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:47:59.227+0000","completedAt":null,"tasks":[]},{"id":"201612121953592YSlf","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:53:59.278+0000","completedAt":null,"tasks":[]},{"id":"20161212192359aTWwb","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:23:59.026+0000","completedAt":null,"tasks":[]},{"id":"20161212193359ppqF0","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:33:59.101+0000","completedAt":null,"tasks":[]},{"id":"20161212195959c71yS","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:59:59.328+0000","completedAt":null,"tasks":[]},{"id":"20161212193959sxqfe","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:39:59.150+0000","completedAt":null,"tasks":[]},{"id":"20161212200359noaiJ","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T20:03:59.357+0000","completedAt":null,"tasks":[]},{"id":"20161212191559OFafE","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:15:59.967+0000","completedAt":null,"tasks":[]},{"id":"20161212191759KtwAA","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:17:59.977+0000","completedAt":null,"tasks":[]},{"id":"20161212192559reYwe","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:25:59.049+0000","completedAt":null,"tasks":[]},{"id":"201612121949597k1qV","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:49:59.237+0000","completedAt":null,"tasks":[]},{"id":"20161212193759AsLk8","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:37:59.138+0000","completedAt":null,"tasks":[]},{"id":"20161212200159eu7AS","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T20:01:59.349+0000","completedAt":null,"tasks":[]},{"id":"20161212190959uY5Gv","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:09:59.921+0000","completedAt":null,"tasks":[]},{"id":"201612121857593BUlr","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T18:57:59.827+0000","completedAt":null,"tasks":[]},{"id":"20161212195159i31hd","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:51:59.257+0000","completedAt":null,"tasks":[]},{"id":"20161212200559UpGSC","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T20:05:59.377+0000","completedAt":null,"tasks":[]},{"id":"20161212193559jmPF7","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:35:59.118+0000","completedAt":null,"tasks":[]},{"id":"20161212195759NZBoi","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:57:59.306+0000","completedAt":null,"tasks":[]},{"id":"20161212191359AninT","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:13:59.957+0000","completedAt":null,"tasks":[]},{"id":"201612121859592GqyE","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T18:59:59.847+0000","completedAt":null,"tasks":[]},{"id":"20161212193159Aqg3w","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:31:59.077+0000","completedAt":null,"tasks":[]},{"id":"20161212190559BKt7A","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:05:59.890+0000","completedAt":null,"tasks":[]},{"id":"20161212195559rJPNh","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:55:59.302+0000","completedAt":null,"tasks":[]},{"id":"20161212190359pVtaU","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:03:59.878+0000","completedAt":null,"tasks":[]},{"id":"20161212201359DtsgH","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T20:13:59.449+0000","completedAt":null,"tasks":[]},{"id":"20161212194359ikn8A","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:43:59.190+0000","completedAt":null,"tasks":[]},{"id":"20161212191959REe2G","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:19:59.997+0000","completedAt":null,"tasks":[]},{"id":"20161212201159yMUic","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T20:11:59.428+0000","completedAt":null,"tasks":[]},{"id":"201612121853596eV58","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T18:53:59.798+0000","completedAt":null,"tasks":[]},{"id":"20161212192959GWEPL","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:29:59.067+0000","completedAt":null,"tasks":[]},{"id":"20161212194559LkryV","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:45:59.213+0000","completedAt":null,"tasks":[]},{"id":"20161212190159Bji8F","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:01:59.858+0000","completedAt":null,"tasks":[]},{"id":"20161212191159WnbI6","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:11:59.937+0000","completedAt":null,"tasks":[]},{"id":"20161212200959c7BoQ","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T20:09:59.407+0000","completedAt":null,"tasks":[]},{"id":"20161212194159Er4ln","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:41:59.167+0000","completedAt":null,"tasks":[]},{"id":"20161212185559b21Bi","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T18:55:59.807+0000","completedAt":null,"tasks":[]},{"id":"201612121907597DflY","jobId":"foo.bar","status":"INITIAL","createdAt":"2016-12-12T19:07:59.897+0000","completedAt":null,"tasks":[]}],"history":{"successCount":19,"failureCount":1,"lastSuccessAt":"2016-12-12T18:02:00.251+0000","lastFailureAt":"2016-12-12T17:36:00.409+0000","successfulFinishedRuns":[{"id":"20161212180159qfl4J","createdAt":"2016-12-12T18:01:59.335+0000","finishedAt":"2016-12-12T18:02:00.251+0000"},{"id":"201612121759592WzzE","createdAt":"2016-12-12T17:59:59.314+0000","finishedAt":"2016-12-12T18:00:00.271+0000"},{"id":"20161212175759EuhhS","createdAt":"2016-12-12T17:57:59.294+0000","finishedAt":"2016-12-12T17:58:00.173+0000"},{"id":"20161212175559OsNh4","createdAt":"2016-12-12T17:55:59.277+0000","finishedAt":"2016-12-12T17:56:00.191+0000"},{"id":"201612121753593PwVS","createdAt":"2016-12-12T17:53:59.257+0000","finishedAt":"2016-12-12T17:54:00.387+0000"},{"id":"20161212175159quwN5","createdAt":"2016-12-12T17:51:59.245+0000","finishedAt":"2016-12-12T17:52:00.164+0000"},{"id":"20161212174959jbcuu","createdAt":"2016-12-12T17:49:59.230+0000","finishedAt":"2016-12-12T17:50:00.165+0000"},{"id":"20161212174759EuLBH","createdAt":"2016-12-12T17:47:59.207+0000","finishedAt":"2016-12-12T17:48:00.194+0000"},{"id":"20161212174559Yk4Cc","createdAt":"2016-12-12T17:45:59.186+0000","finishedAt":"2016-12-12T17:46:00.078+0000"},{"id":"20161212174359IdEUB","createdAt":"2016-12-12T17:43:59.164+0000","finishedAt":"2016-12-12T17:44:00.098+0000"},{"id":"20161212174159QbXKn","createdAt":"2016-12-12T17:41:59.154+0000","finishedAt":"2016-12-12T17:42:00.037+0000"},{"id":"201612121739593nv1S","createdAt":"2016-12-12T17:39:59.134+0000","finishedAt":"2016-12-12T17:40:00.114+0000"},{"id":"20161212173759jMH95","createdAt":"2016-12-12T17:37:59.113+0000","finishedAt":"2016-12-12T17:37:59.950+0000"},{"id":"201612121733597QS6t","createdAt":"2016-12-12T17:33:59.743+0000","finishedAt":"2016-12-12T17:34:00.561+0000"},{"id":"201612121731598j283","createdAt":"2016-12-12T17:31:59.733+0000","finishedAt":"2016-12-12T17:32:00.648+0000"},{"id":"20161212172959BqBYh","createdAt":"2016-12-12T17:29:59.715+0000","finishedAt":"2016-12-12T17:30:00.546+0000"},{"id":"20161212172759WGAh2","createdAt":"2016-12-12T17:27:59.694+0000","finishedAt":"2016-12-12T17:28:00.703+0000"},{"id":"20161212172559lbS9Z","createdAt":"2016-12-12T17:25:59.689+0000","finishedAt":"2016-12-12T17:26:00.626+0000"},{"id":"20161212172359E8E0G","createdAt":"2016-12-12T17:23:59.665+0000","finishedAt":"2016-12-12T17:24:00.587+0000"}],"failedFinishedRuns":[{"id":"20161212173559asQpr","createdAt":"2016-12-12T17:35:59.483+0000","finishedAt":"2016-12-12T17:36:00.409+0000"}]},"historySummary":{"successCount":19,"failureCount":1,"lastSuccessAt":"2016-12-12T18:02:00.251+0000","lastFailureAt":"2016-12-12T17:36:00.409+0000"}}
					`),
				))
		})

		It("Makes a request to get all jobs", func() {
			result,err  := client.RunLs("foo.bar")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(server.ReceivedRequests()).To(HaveLen(2))
			fmt.Printf("%+v\n",result)
		})
	})

	Describe("DeleteJob", func() {
		var (
			jobName = "job.with.arguments"
		)

		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/v1/jobs/" + jobName),
					ghttp.RespondWithJSONEncoded(http.StatusOK, allJobs[0]),
				),
			)
		})

		It("Makes the delete request", func() {
			rez, err := client.DeleteJob(jobName)
			Expect(err).ShouldNot(HaveOccurred())
			tt := rez.(Job)
			Expect(tt.ID_).To(Equal(allJobs[0].ID_))
			Expect((*tt.Labels_)["Location"]).To(Equal((*allJobs[0].Labels())["Location"]))
			Expect((*tt.Labels_)["Owner"]).To(Equal((*allJobs[0].Labels())["Owner"]))
			//			Expect(tt).Should(Equal(allJobs[0])) //To(Succeed())
			Expect(server.ReceivedRequests()).To(HaveLen(2))
		})
	})

	Describe("StartJob", func() {
		var (
			job_without_arguments = "job.without.arguments"
		)

		Context("Starting a job", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf("/v1/jobs/%s/runs", job_without_arguments), ""),
						ghttp.RespondWithJSONEncoded(http.StatusOK, status),
					),
				)
			})

			It("Makes the start request", func() {
				st, err := client.RunStartJob(job_without_arguments)
				Expect(err).ShouldNot(HaveOccurred())

				_, found := st.(JobStatus)
				Expect(found).To(Equal(true))

				Expect(server.ReceivedRequests()).To(HaveLen(2))
			})
		})

	})

	Describe("AddScheduledJob", func() {
		var (
			some_job = "job.with.arguments"
		)

		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf("/v1/jobs/%s", some_job)),
					ghttp.VerifyJSONRepresenting(Job{}),
					ghttp.RespondWithJSONEncoded(http.StatusCreated, allJobs[0]),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf("/v1/jobs/%s/schedules", some_job)),
					ghttp.VerifyJSONRepresenting(Schedule{}),
					ghttp.RespondWith(http.StatusCreated, sched),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf("/v1/jobs/%s/runs", some_job)),
					ghttp.RespondWith(http.StatusCreated, status),
				),
			)
		})
		/*
				Describe("RunOnceNowJob", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("POST", "/scheduler/iso8601"),
								ghttp.VerifyJSON(`{"name":"","command":"","epsilon":"PT10M","schedule":"R1//PT2M"}`),
								ghttp.RespondWith(http.StatusOK, nil),
							),
						)
					})

					It("Schedules a job to run once, and start immediately", func() {
						job := Job{}
						Expect(client.RunOnceNowJob(&job)).To(Succeed())
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Describe("FormatSchedule", func() {
					It("Returns a properly formatted time string", func() {
						startTime := time.Date(2015, time.May, 26, 15, 0, 0, 0, time.UTC)
						interval := "P10M"
						reps := "R10"
						expectedOutput := "R10/2015-05-26T15:00:00Z/P10M"

						Expect(FormatSchedule(startTime, interval, reps)).To(Equal(expectedOutput))
					})

					It("Works with a zero time", func() {
						startTime := *new(time.Time)
						interval := "P10M"
						reps := "R10"
						expectedOutput := "R10//P10M"

						Expect(FormatSchedule(startTime, interval, reps)).To(Equal(expectedOutput))
					})

					It("Errors if interval does not start with a P", func() {
						startTime := new(time.Time)
						interval := "10M"
						reps := "R10"

						schedule, err := FormatSchedule(*startTime, interval, reps)
						Expect(schedule).To(Equal(""))
						Expect(err).To(MatchError("Interval string not formatted correctly"))
					})

					It("Errors if reps do not start with R", func() {
						startTime := new(time.Time)
						interval := "P10M"
						reps := "10"

						schedule, err := FormatSchedule(*startTime, interval, reps)
						Expect(schedule).To(Equal(""))
						Expect(err).To(MatchError("Repetitions string not formatted correctly"))
					})
				})
		*/
	})
})
