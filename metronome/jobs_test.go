package metronome_test

import (
	"net/http"
	"time"

	. "github.com/adobe-platform/go-metronome/metronome"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	ghttp "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Jobs", func() {
	var (
		config_stub Config
		client      Metronome
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
							"name":"dockerjob",
							"command":"while sleep 10; do date -u +%T; done",
							"shell":true,
							"epsilon":"PT60S",
							"executor":"",
							"executorFlags":"",
							"retries":2,
							"owner":"",
							"ownerName":"",
							"description":"",
							"async":false,
							"successCount":190,
							"errorCount":3,
							"lastSuccess":"2014-03-08T16:57:17.507Z",
							"lastError":"2014-03-01T00:10:15.957Z",
							"cpus":0.5,
							"disk":256.0,
							"mem":512.0,
							"disabled":false,
							"softError":false,
							"dataProcessingJobType":false,
							"errorsSinceLastSuccess":0,
							"uris":[],
							"environmentVariables":[],
							"arguments":[],
							"highPriority":false,
							"runAsUser":"root",
							"container":{
								"type":"docker",
								"image":"libmesos/ubuntu",
								"network":"HOST",
								"volumes":[]
							},
							"schedule":"R/2015-05-21T18:14:00.000Z/PT2M",
							"scheduleTimeZone":""
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
			Expect(jobs).To(Equal(&Jobs{
				Job{
					Name:                 "dockerjob",
					Command:              "while sleep 10; do date -u +%T; done",
					Shell:                true,
					Epsilon:              "PT60S",
					Executor:             "",
					ExecutorFlags:        "",
					Retries:              2,
					Owner:                "",
					Async:                false,
					SuccessCount:         190,
					ErrorCount:           3,
					LastSuccess:          "2014-03-08T16:57:17.507Z",
					LastError:            "2014-03-01T00:10:15.957Z",
					CPUs:                 0.5,
					Disk:                 256,
					Mem:                  512,
					Disabled:             false,
					URIs:                 []string{},
					Schedule:             "R/2015-05-21T18:14:00.000Z/PT2M",
					EnvironmentVariables: []map[string]string{},
					Arguments:            []string{},
					RunAsUser:            "root",
					Container: &Container{
						Type:    "docker",
						Image:   "libmesos/ubuntu",
						Network: "HOST",
						Volumes: []map[string]string{},
					},
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
					ghttp.VerifyRequest("DELETE", "/scheduler/job/"+jobName),
					ghttp.RespondWith(http.StatusOK, nil),
				),
			)
		})

		It("Makes the delete request", func() {
			Expect(client.DeleteJob(jobName)).To(Succeed())
			Expect(server.ReceivedRequests()).To(HaveLen(2))
		})
	})

	Describe("DeleteJobTasks", func() {
		var (
			jobName = "fake_job"
		)

		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/scheduler/task/kill/"+jobName),
					ghttp.RespondWith(http.StatusOK, nil),
				),
			)
		})

		It("Makes the delete request", func() {
			Expect(client.DeleteJobTasks(jobName)).To(Succeed())
			Expect(server.ReceivedRequests()).To(HaveLen(2))
		})
	})

	Describe("StartJob", func() {
		var (
			jobName = "fake_job"
		)

		Context("Starting a job with no arguments", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/scheduler/job/"+jobName, ""),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("Makes the start request", func() {
				Expect(client.StartJob(jobName, nil)).To(Succeed())
				Expect(server.ReceivedRequests()).To(HaveLen(2))
			})
		})

		Context("Starting a job with arguments", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/scheduler/job/"+jobName, "arg1=value1&arg2=value2"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
			})

			It("Can pass arguments to the start job request", func() {
				args := map[string]string{
					"arg1": "value1",
					"arg2": "value2",
				}

				Expect(client.StartJob(jobName, args)).To(Succeed())
				Expect(server.ReceivedRequests()).To(HaveLen(2))
			})
		})
	})

	Describe("AddScheduledJob", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/scheduler/iso8601"),
					ghttp.VerifyJSONRepresenting(Job{}),
					ghttp.RespondWith(http.StatusOK, nil),
				),
			)
		})

		It("Makes the request", func() {
			job := Job{}
			Expect(client.AddScheduledJob(&job)).To(Succeed())
			Expect(server.ReceivedRequests()).To(HaveLen(2))
		})
	})

	Describe("AddDependentJob", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/scheduler/dependency"),
					ghttp.VerifyJSONRepresenting(Job{}),
					ghttp.RespondWith(http.StatusOK, nil),
				),
			)
		})

		It("Makes the request", func() {
			job := Job{}
			Expect(client.AddDependentJob(&job)).To(Succeed())
			Expect(server.ReceivedRequests()).To(HaveLen(2))
		})
	})

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
})
