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
						    "location": "olympus",
						    "owner": "zeus"
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
						    "location": "olympus",
						    "owner": "zeus"
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
			Expect((*jobs)[0]).To(Equal(Job{ID: "job.with.arguments",
				Description: "Job with arguments",
				Labels: &Labels{
					Location: "olympus",
					Owner: "zeus",
				},
				Run: &Run{
					Artifacts: []Artifact{
						Artifact{URI: "http://foo.test.com/application.zip", Extract: true, Executable :true, Cache: false},
					},
					Cmd: "nuke --dry --master local",
					Args:[]string{
						"nuke",
						"--dry",
						"--master",
						"local",
					},
					Cpus: 1.5,
					Mem: 128,
					Disk: 32,
					Docker : &Docker{
						Image: "foo/bla:test",
					},
					Env:  map[string]string{
						"MON": "test",
						"CONNECT": "direct",
					},
					MaxLaunchDelay: 3600,
					Placement: &Placement{
						Constraints: []Constraint{
							Constraint{Attribute: "rack", Operator: EQ, Value: "rack-2"} },

					},
					Restart: &Restart{
						ActiveDeadlineSeconds: 120, Policy: "NEVER",

					},
					User: "root",
					Volumes: []Volume{
						Volume{Mode:RW, HostPath:"/etc/guest", ContainerPath: "/mnt/test" },
					},
				},
			}))

			Expect((*jobs)[1]).To(Equal(
				Job{Description: "Job without arguments",
					ID: "job.without.arguments",
					Labels: &Labels{
						Location: "olympus",
						Owner: "zeus",
					},
					Run: &Run{Cmd: "/usr/local/bin/dcos-tests --debug --term-wait 20 --http-addr :8095",
						Cpus: 0.5,
						Disk: 128,
						Docker: &Docker{
							Image: "f4tq/dcos-tests:v0.31",
						},
						Env:   map[string]string{
							"MON": "test",
							"CONNECT": "direct",
						},
						MaxLaunchDelay: 3600,
						Mem: 32,
						Restart: &Restart{
							ActiveDeadlineSeconds: 120,
							Policy: "NEVER",
						},
						User: "root",
					},

				}))

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
			Expect(tt.ID).To(Equal(allJobs[0].ID))
			Expect(tt.Labels.Location).To(Equal(allJobs[0].GetLabels().Location))
			Expect(tt.Labels.Owner).To(Equal(allJobs[0].GetLabels().Owner))
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
				st, err := client.StartJob(job_without_arguments)
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
