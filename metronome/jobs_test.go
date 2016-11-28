package metronome_test

import (
	"net/http"
	//"time"

	. "github.com/adobe-platform/go-metronome/metronome"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	ghttp "github.com/onsi/gomega/ghttp"
	"fmt"
)

var _ = Describe("Jobs", func() {
	var (
		config_stub Config
		client Metronome
		server      *ghttp.Server
	)

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
					ghttp.RespondWith(http.StatusOK, `[
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
					]`),
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
					Location: "olympus",
					Owner: "zeus",
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
						Location: "olympus",
						Owner: "zeus",
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

	Describe("DeleteJob", func() {
		var (
			jobName = "fake_job"
		)

		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/v1/jobs/" + jobName),
					ghttp.RespondWith(http.StatusOK, nil),
				),
			)
		})

		It("Makes the delete request", func() {
			Expect(client.DeleteJob(jobName)).To(Succeed())
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
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("Makes the start request", func() {
				Expect(client.StartJob(job_without_arguments)).To(Succeed())
				Expect(server.ReceivedRequests()).To(HaveLen(2))
			})
		})

	})

	Describe("AddScheduledJob", func() {
		var (

			some_job = "some_job"
		)

		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf("/v1/jobs/%s",some_job)),
					ghttp.VerifyJSONRepresenting(Job{}),
					ghttp.RespondWith(http.StatusOK, nil),
				),
			)
		})

		It("Makes the request", func() {
			job :=Job{ID_: some_job,
				Description_: "Job with arguments",
				Labels_: &Labels{
					Location: "olympus",
					Owner: "zeus",
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
				}}

			now, _ :=ImmediateSchedule()
			Expect(client.AddScheduledJob(&job, now)).To(Succeed())
			Expect(server.ReceivedRequests()).To(HaveLen(2))
		})
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
