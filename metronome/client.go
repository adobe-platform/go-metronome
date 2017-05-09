package metronome

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/behance/go-logrus"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
	"time"
)

// Constants to represent HTTP verbs
const (
	HTTPGet    = "GET"
	HTTPPut    = "PUT"
	HTTPDelete = "DELETE"
	HTTPPost   = "POST"
)

// Metronome represents the client interface for interacting with the metronome API
type Metronome interface {
	// POST /v1/jobs
	CreateJob(*Job) (*Job, error)
	// DELETE /v1/jobs/$jobId
	DeleteJob(jobID string) (interface{}, error)
	// GET /v1/jobs/$jobId
	GetJob(jobID string) (*Job, error)
	// GET /v1/jobs
	Jobs() (*[]Job, error)
	// PUT /v1/jobs/$jobId
	UpdateJob(jobID string, job *Job) (interface{}, error)
	//
	// schedules
	// GET /v1/jobs/$jobId/runs
	// Technically, this rev of Runs() is a hack to get functionality from the undocumented api
	//   - since is milliseconds from epoch

	Runs(jobID string, statusSince int64) (*Job, error)
	// POST /v1/jobs/$jobId/runs
	StartJob(jobID string) (interface{}, error)
	// GET /v1/jobs/$jobId/runs/$runId
	StatusJob(jobID string, runID string) (*JobStatus, error)
	// POST /v1/jobs/$jobId/runs/$runId/action/stop
	StopJob(jobID string, runID string) (interface{}, error)

	//
	// Schedules
	//
	// POST /v1/jobs/$jobId/schedules
	CreateSchedule(jobID string, new *Schedule) (interface{}, error)
	// GET /v1/jobs/$jobId/schedules/$scheduleId
	GetSchedule(jobID string, schedID string) (*Schedule, error)
	// GET /v1/jobs/$jobId/schedules
	Schedules(jobID string) (*[]Schedule, error)
	// DELETE /v1/jobs/$jobId/schedules/$scheduleId
	DeleteSchedule(jobID string, schedID string) (interface{}, error)
	// PUT /v1/jobs/$jobId/schedules/$scheduleId
	UpdateSchedule(jobID string, schedID string, sched *Schedule) (interface{}, error)

	//  GET  /v1/metrics
	Metrics() (interface{}, error)
	//  GET /v1/ping
	Ping() (*string, error)
}

// TwentyFourHoursAgo - return time 24 hours ago
func TwentyFourHoursAgo() int64 {
	return time.Now().UnixNano()/int64(time.Millisecond) - 24*3600000
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
	log.Debugf("NewClient started %+v", config)
	var err error
	client.url, err = url.Parse(config.URL)
	if err != nil {
		return nil, err
	}
	client.config = config
	var PTransport http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.AllowUnverifiedTls,
		},
	}

	client.http = &http.Client{
		Timeout:   (time.Duration(config.RequestTimeout) * time.Second),
		Transport: PTransport,
	}
	// Verify you can reach metronome
	_, err = client.Jobs()
	if err != nil {
		return nil, errors.New("Could not reach metronome cluster: " + err.Error())
	}

	return client, nil
}

func (client *Client) apiGet(uri string, queryParams map[string][]string, result interface{}) (status int, err error) {
	return client.apiCall(HTTPGet, uri, queryParams, "", result)
}

func (client *Client) apiDelete(uri string, queryParams map[string][]string, result interface{}) (status int, err error) {
	return client.apiCall(HTTPDelete, uri, queryParams, "", result)

}

func (client *Client) apiPut(uri string, queryParams map[string][]string, putData interface{}, result interface{}) (status int, err error) {

	var putDataString []byte
	if putData != nil {
		putDataString, err = json.Marshal(putData)
		log.Debugf("PUT %s", string(putDataString))
	}
	return client.apiCall(HTTPPut, uri, queryParams, string(putDataString), result)
}

func (client *Client) apiPost(uri string, queryParams map[string][]string, postData interface{}, result interface{}) (status int, err error) {
	//postDataString, err := json.Marshal(postData)
	postDataString := new(bytes.Buffer)
	enc := json.NewEncoder(postDataString)
	enc.SetEscapeHTML(false)
	err = enc.Encode(postData)

	if err != nil {
		return http.StatusBadRequest, err
	}

	return client.apiCall(HTTPPost, uri, queryParams, postDataString.String(), result)

}

func (client *Client) apiCall(method string, uri string, queryParams map[string][]string, body string, result interface{}) (int, error) {
	log.Debugf("apiCall ... method: %v url: %v queryParams: %+v", method, uri, queryParams)

	url, _ := client.buildURL(uri, queryParams)
	status, response, err := client.httpCall(method, url, body)

	if err != nil {
		return 0, err
	}
	log.Debugf("%s result status: %+v", uri, response.Status)
	log.Debugf("Headers: %+v", response.Header)
	if response.ContentLength > 0 {
		ct := response.Header["Content-Type"]
		log.Debugf("content-type: %s", ct)
		switch ct[0] {
		case "application/json":
			var msg json.RawMessage
			err = json.NewDecoder(response.Body).Decode(&msg)
			// decode as a raw json message which will fail if the message isn't good json
			if err == nil {
				switch result.(type) {
				case json.RawMessage:
					tt := result.(*json.RawMessage)
					*tt = msg
					return status, nil
				default:
					err = json.Unmarshal(msg, result)
					if err != nil || status >= 400 {
						//== http.StatusUnprocessableEntity {
						// metronome returns json error messages.  panic if so.
						bb := new(bytes.Buffer)
						fmt.Fprintf(bb, string(msg))
						return status, errors.New(string(bb.Bytes()))
					}
					log.Debugf("method %s uri: %s status: %d result type: %T", method, uri, status, result)
				}
			} else {
				return status, err
			}

		case "text/plain; charset=utf-8":
			htmlData, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return status, err
			}
			v := result.(*string)
			*v = string(htmlData)

		default:
			return status, fmt.Errorf("Unknown content-type %s", ct[0])
		}
	}

	// TODO: Handle error status codes
	if status < 200 || status > 299 {
		return status, errors.New(response.Status)
	}
	return status, nil
}
func (client *Client) buildURL(reqPath string, queryParams map[string][]string) (*url.URL, error) {
	// make copy of client url
	base := *client.url

	query := base.Query()
	log.Debugf("client.url.params %+v ; queryParams: %+v; client.config.URL: %+v base.url: %+v", query, queryParams, client.config.URL, base)
	master, _ := url.Parse(client.config.URL)
	prefix := master.Path
	for k, vl := range queryParams {
		for _, val := range vl {
			query.Add(k, val)
		}
	}
	base.RawQuery = query.Encode()

	base.Path = path.Join(prefix, reqPath)
	return &base, nil
}

func (client *Client) applyRequestHeaders(request *http.Request) {
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", "application/json")
	if client.config.User != "" && client.config.Pw != "" {
		request.SetBasicAuth(client.config.User, client.config.Pw)
	}
	if client.config.AuthToken != "" {
		request.Header.Add("Authorization", client.config.AuthToken)
	}
}

func (client *Client) newRequest(method string, url *url.URL, body string) (*http.Request, error) {
	request, err := http.NewRequest(method, url.String(), strings.NewReader(body))

	if err != nil {
		return nil, err
	}

	client.applyRequestHeaders(request)
	if client.config.Debug {
		if dump, err := httputil.DumpRequest(request, true); err != nil {
			log.Infof(string(dump))
		}
	}
	return request, nil
}

func (client *Client) httpCall(method string, url *url.URL, body string) (int, *http.Response, error) {
	request, err := client.newRequest(method, url, body)

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
	log.Infof(message+"\n", args...)
}
