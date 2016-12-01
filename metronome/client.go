package metronome

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"io/ioutil"
	//"log"
	"github.com/Sirupsen/logrus"
)

// Constants to represent HTTP verbs
const (
	HTTPGet = "GET"
	HTTPPut = "PUT"
	HTTPDelete = "DELETE"
	HTTPPost = "POST"
)

// Metronome is a client that can interact with the metronome API
type MetronomeChronos interface {
	Jobs() (*[]Job, error)
	DeleteJob(job_id string) error
	StartJob(name string) error
	AddScheduledJob(job *Job, sched *Schedule) error
	RunOnceNowJob(job *Job) error
}

type Metronome interface {
	// POST /v1/jobs
	CreateJob(*Job) (interface{}, error)
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
	Ping() (interface{}, error)
}



// A Client can make http requests
type Client struct {
	url    *url.URL
	config Config
	http   *http.Client
}

// NewClient returns a new  client, initialzed with the provided config
func NewClient(config Config) (Metronome, error) {
	client := new(Client)
	logrus.Debugf("NewClient started\n")
	var err error
	client.url, err = url.Parse(config.URL)
	if err != nil {
		return nil, err
	}
	client.config = config
	client.http = &http.Client{
		Timeout: (time.Duration(config.RequestTimeout) * time.Second),
	}
	// Verify you can reach metronome
	_, err = client.Jobs()
	if err != nil {
		return nil, errors.New("Could not reach metronome cluster: " + err.Error())
	}

	return client, nil
}

func (client *Client) apiGet(uri string, queryParams map[string]string, result interface{}) error {
	_, err := client.apiCall(HTTPGet, uri, queryParams, "", result)
	return err
}

func (client *Client) apiDelete(uri string, queryParams map[string]string, result interface{}) error {
	_, err := client.apiCall(HTTPDelete, uri, queryParams, "", result)
	return err
}

func (client *Client) apiPut(uri string, queryParams map[string]string, result interface{}) error {
	_, err := client.apiCall(HTTPPut, uri, queryParams, "", result)
	return err
}

func (client *Client) apiPost(uri string, queryParams map[string]string, postData interface{}, result interface{}) error {
	postDataString, err := json.Marshal(postData)

	if err != nil {
		return err
	}

	_, err = client.apiCall(HTTPPost, uri, queryParams, string(postDataString), result)
	return err
}

func (client *Client) apiCall(method string, uri string, queryParams map[string]string, body string, result interface{}) (int, error) {
	client.buildURL(uri, queryParams)
	status, response, err := client.httpCall(method, body)

	if err != nil {
		return 0, err
	}
	logrus.Debugf("%s result status: %d\n", uri, response.Status)
	logrus.Debugf("Headers: %+v\n\n", response.Header)
	if response.ContentLength > 0 {
		ct := response.Header["Content-Type"]
		logrus.Debugf("content-type: %s\n",ct)
		switch ct[0] {
		case "application/json":
			err = json.NewDecoder(response.Body).Decode(result)
			if err != nil {
				return status, err
			}
		case "text/plain; charset=utf-8":
			if htmlData, err := ioutil.ReadAll(response.Body); err != nil {
				return status, err
			} else {
				v := result.(*string)
				v2 := string(htmlData)
				v = &v2
				// compiler see v as unused
				if v != nil{}
				logrus.Debugf("text bytes: %d result: %+v %+v\n",len(htmlData),htmlData,v)
			}
		default:
			return status,errors.New(fmt.Sprintf("Unknown content-type %s",ct[0]))
		}

	}

	// TODO: Handle error status codes
	if status < 200 || status > 299 {
		return status, errors.New(response.Status)
	}
	return status, nil
}
func (client *Client) buildURL(path string, queryParams map[string]string) {
	query := client.url.Query()
	for k, v := range queryParams {
		query.Add(k, v)
	}
	client.url.RawQuery = query.Encode()

	client.url.Path = path
}

// TODO: think about pulling out a Request struct/object/thing
func (client *Client) applyRequestHeaders(request *http.Request) {
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", "application/json")
}

func (client *Client) newRequest(method string, body string) (*http.Request, error) {
	request, err := http.NewRequest(method, client.url.String(), strings.NewReader(body))

	if err != nil {
		return nil, err
	}

	client.applyRequestHeaders(request)
	return request, nil
}

func (client *Client) httpCall(method string, body string) (int, *http.Response, error) {
	request, err := client.newRequest(method, body)

	if err != nil {
		return 0, nil, err
	}

	response, err := client.http.Do(request)

	if err != nil {
		return 0, nil, err
	}

	return response.StatusCode, response, nil
}

// TODO: this better
func (client *Client) log(message string, args ...interface{}) {
	fmt.Printf(message + "\n", args...)
}
