package metronome

import (
	"errors"
	"fmt"
	"strings"
	"time"
	duration "github.com/ChannelMeter/iso8601duration"
	"encoding/json"
	"regexp"
	"strconv"
	log "github.com/behance/go-logrus"
	"net/http"
)
// CreateJob - create a metronome job.  returns the job or an error
func (client *Client) CreateJob(job *Job) (*Job, error) {
	var reply Job
	if _, err := client.apiPost(MetronomeAPIJobCreate, nil, job, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

// DeleteJob - deletes a job by calling metronome api
// DELETE /v1/jobs/$jobId
func (client *Client)  DeleteJob(jobID string) (interface{}, error) {
	var msg Job //json.RawMessage
	_, err := client.apiDelete(fmt.Sprintf(MetronomeAPIJobDelete, jobID), nil, &msg)
	if err != nil {
		return nil, err
	}
	return msg, err

}
// GetJob - Gets a job by calling metronome api
// GET /v1/jobs/$jobId
func (client *Client) GetJob(jobID string) (*Job, error) {
	var job Job
	queryParams := map[string][]string{
		"embed" : {
			"historySummary",
			"activeRuns",
			"schedules",
		},
	}
	_, err := client.apiGet(fmt.Sprintf(MetronomeAPIJobGet, jobID), queryParams, &job)
	if err != nil {
		return nil, err
	}
	return &job, err

}
// Jobs - get a list of all jobs by calling metronome api
// GET /v1/jobs
func (client *Client)  Jobs() (*[]Job, error) {
	//	jobs := new(Jobs)
	jobs := make([]Job, 0, 0)
	queryParams := map[string][]string{
		"embed" : {
			"historySummary",
			"activeRuns",
		},
	}

	_, err := client.apiGet(MetronomeAPIJobList, queryParams, &jobs)

	if err != nil {
		return nil, err
	}
	return &jobs, nil
}


// UpdateJob - given jobID and new job structure, replace an existing job by calling metronome api
// PUT /v1/jobs/$jobId
func (client *Client) UpdateJob(jobID string, job *Job) (interface{}, error) {
	var msg json.RawMessage
	_, err := client.apiPut(fmt.Sprintf(MetronomeAPIJobUpdate, jobID), nil, job, &msg)
			if err != nil {
		bbb, err2 := json.Marshal(msg)
				if err2 != nil {
					return nil, fmt.Errorf("JobUpdate error %s\n\tAnd %s", err.Error(), err2.Error())
		}
		return nil, fmt.Errorf("JobUpdate error %s\n%s", err, string(bbb))

	}
	return &msg, nil
}

// Runs - get all the 'runs' of a given job
// GET /v1/jobs/$jobId/runs
func (client *Client) Runs(jobID string, since int64) (*Job, error) {
	//jobs := make([]JobStatus, 0, 0)
	//jobs := make([]Job, 0, 0)
	var jobs Job
	queryParams := map[string][]string{
		"_timestamp": {
			strconv.FormatInt(since , 10),
//			strconv.FormatInt(time.Now().UnixNano() / int64(time.Millisecond) - 24 * 3600000, 10),
		},
		"embed" : {
			"history",
			"historySummary",
			"activeRuns",
			"schedules",
		},
	}
	// lame hidden parameters are only reachable via /v1/jobs/$jobId with queryParams
	_, err := client.apiGet(fmt.Sprintf(MetronomeAPIJobGet, jobID), queryParams, &jobs)
	if err != nil {
		return nil, err
	}
	return &jobs, nil
}
// RunLs  - list running jobs - standard
func (client *Client) RunLs(jobID string) (*[]JobStatus, error) {
	jobs := make([]JobStatus, 0, 0)

	_, err := client.apiGet(fmt.Sprintf(MetronomeAPIJobRunList, jobID), nil, &jobs)

	if err != nil {
		return nil, err
	}
	return &jobs, nil
}
// StartJob - starts a metronome job.  Implies that CreateJob was already called.
// POST /v1/jobs/$jobId/runs
func (client *Client) StartJob(jobID string) (interface{}, error) {
	var msg JobStatus
	if _, err := client.apiPost(fmt.Sprintf(MetronomeAPIJobRunStart, jobID), nil, jobID, &msg); err != nil {
		return nil, err
	}
	return msg, nil
}
// StatusJob - get a job status
// GET /v1/jobs/$jobId/runs/$runId
func (client *Client)  StatusJob(jobID string, runID string) (*JobStatus, error) {
	var job JobStatus

	_, err := client.apiGet(fmt.Sprintf(MetronomeAPIJobRunStatus, jobID, runID), nil, &job)
	if err != nil {
		return nil, err
	}
	return &job, err

}

// StopJob - stop a running job.  returns and error on failure
// POST /v1/jobs/$jobId/runs/$runId/action/stop
func (client *Client) StopJob(jobID string, runID string) (interface{}, error) {
	var msg json.RawMessage
	if _, err := client.apiPost(fmt.Sprintf(MetronomeAPIJobRunStop, jobID, runID), nil, jobID, &msg); err != nil {
		return nil, err
	}
	return msg, nil
}

//
// Schedules
//

// CreateSchedule - assign a schedule to a job
// POST /v1/jobs/$jobId/schedules
func (client *Client) CreateSchedule(jobID string, sched *Schedule) (interface{}, error) {
	var msg Schedule //json.RawMessage
	log.Debugf("client.JobScheduleCreate %s\n", jobID)
	if _, err := client.apiPost(fmt.Sprintf(MetronomeAPIJobScheduleCreate, jobID), nil, sched, &msg); err != nil {
		return nil, err
	}
	return msg, nil

}

// GetSchedule - get a schedule associated with a job
// GET /v1/jobs/$jobId/schedules/$scheduleId
func (client *Client) GetSchedule(jobID string, schedID string) (*Schedule, error) {
	var sched Schedule

	_, err := client.apiGet(fmt.Sprintf(MetronomeAPIJobScheduleStatus, jobID, schedID), nil, &sched)
		if err != nil {
		return nil, err
	}
	fmt.Printf("sched: %+v\n", sched)
	return &sched, err

}
// Schedules - get all schedules
// GET /v1/jobs/$jobId/schedules
func (client *Client) Schedules(jobID string) (*[]Schedule, error) {
	scheds := make([]Schedule, 0, 0)

	_, err := client.apiGet(fmt.Sprintf(MetronomeAPIJobScheduleList, jobID), nil, &scheds)

	if err != nil {
		return nil, err
	}
	return &scheds, nil
}

// DeleteSchedule - delete a schedule
// DELETE /v1/jobs/$jobId/schedules/$scheduleId
func (client *Client) DeleteSchedule(jobID string, schedID string) (interface{}, error) {
	var msg json.RawMessage
	status, err := client.apiDelete(fmt.Sprintf(MetronomeAPIJobScheduleDelete, jobID, schedID), nil, &msg)
	if err != nil {
		return nil, err
	}
	if len([]byte(msg)) == 0 {
		return http.StatusText(status), nil
	}
	return msg, err

}
// UpdateSchedule - update an existing schedule associated with a job
// PUT /v1/jobs/$jobId/schedules/$scheduleId
func (client *Client) UpdateSchedule(jobID string, schedID string, sched *Schedule) (interface{}, error) {
	var msg json.RawMessage
	_, err := client.apiPut(fmt.Sprintf(MetronomeAPIJobScheduleUpdate, jobID, schedID), nil, sched, &msg)
	if err != nil {
		bbb, err2 := json.Marshal(msg)
		if err2 != nil {
			return nil, fmt.Errorf("JobScheduleUpdate multiple errors: %s / %s", err.Error(), err2.Error())
		}
		return nil, fmt.Errorf("JobScheduleUpdate multiple error %s / %s", err, string(bbb))

	}
	return sched, nil

}
// Metrics - returns metrics from the metronome service
//  GET  /v1/metrics
func (client *Client) Metrics() (interface{}, error) {
	msg := json.RawMessage{}
	_, err := client.apiGet(MetronomeAPIMetrics, nil, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, err

}

// Ping - test if the metronome service is running. returns 'pong' on success
//  GET /v1/ping
func (client *Client) Ping() (*string, error) {
	val := new(string)
	msg := (interface{})(val)
	_, err := client.apiGet(MetronomeAPIPing, nil, msg)
	if err != nil {
		return nil, err
	}
	// use Sprintf to reflect the value out.  painful
	retval := fmt.Sprintf("%s", *msg.(*string))
	return &retval, err

}

// RunOnceNowSchedule will return a schedule that starts immediately, runs once,
// and runs every 2 minutes until successful
func RunOnceNowSchedule() string {
	return ImmediateCrontab()
}

func formatTimeString(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.Format(time.RFC3339Nano)
}


///
var (
	repeatRegex = regexp.MustCompile(`R((?P<repeat>\d*))`)
)

// ConvertIso8601ToCron - attempts to convert an iso8601, 3 -part date into a cron repre.  Experimental
//  - only simple cases will work without creating *multiple* schedules
//
func ConvertIso8601ToCron(isoRep string) (string, error) {
	pat := strings.Split(isoRep, "/")
	if len(pat) == 3 {
		interval := pat[2]
		dur := pat[0]
		repeatTimes := 0
		if repeatRegex.MatchString(dur) {
			match := repeatRegex.FindStringSubmatch(dur)
			for i, name := range repeatRegex.SubexpNames() {
				part := match[i]
				if i == 0 || name == "" || part == "" {
					continue
				}
				val, err := strconv.Atoi(part)
				if err != nil {
					return "", err
				}
				switch name {
				case "repeat":
					repeatTimes = val
				default:
					return "", fmt.Errorf("unknown field %s", name)
				}
			}
		} else {
			return "", fmt.Errorf("No repeat pattern")

		}
		tdur, err := duration.FromString(interval)

		if err != nil {
			return "", errors.New("Illegal duration")
		}
		timeT := tdur.ToDuration()
		if repeatTimes != 0 {
			// minute is the smallest scheduling unit for metronome
			slot := int64(timeT)
			if slot < 1 {
				return "", errors.New("Too small a duration")
			} else if slot < 60 {

			}

		} else {

		}

	} else {
		var (
			y, M, d, h, m, s int
		)
		_, err := fmt.Sscanf(time.Now().Format(time.RFC3339), "%d-%d-%dT%d:%d:%dZ", &y, &M, &d, &h, &m, &s)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d %d %d %d * %d", m, h, d, M, y), nil

	}

	return "", errors.New("Unknown error")
}
// ImmediateCrontab - generates a point in time crontab
func ImmediateCrontab() string {
	var (
		y, M, d, h, m, s int
	)
	_, err := fmt.Sscanf(time.Now().Format(time.RFC3339), "%d-%d-%dT%d:%d:%dZ", &y, &M, &d, &h, &m, &s)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d %d %d %d * %d", m, h, d, M, y)
}
// ImmediateSchedule - create a metronome schedule
func ImmediateSchedule() (*Schedule, error) {
	var (
		y, M, d, h, m, s int
		cronstr string
	)
	_, err := fmt.Sscanf(time.Now().Format(time.RFC3339), "%d-%d-%dT%d:%d:%dZ", &y, &M, &d, &h, &m, &s)
	if err != nil {
		return nil, err
	}
	cronstr = fmt.Sprintf("%d %d %d %d * %d", m, h, d, M, y)

	sched := &Schedule{
		ID:  fmt.Sprintf("%d%d%d%d%d%d ", y, M, d, h, m, s), //"everyminute",
		Cron: cronstr, //"cron": "* * * * *",
		ConcurrencyPolicy: "ALLOW",
		Enabled: true,
		StartingDeadlineSeconds:60,
		Timezone: "GMT",
	}
	return sched, nil
}
