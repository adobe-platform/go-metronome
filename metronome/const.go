package metronome

// Constants defining the various metronome endpoints
const (
	// Job model

	// POST /v1/jobs
	MetronomeAPIJobCreate = "/v1/jobs/%s"
	// DELETE /v1/jobs/$jobId
	MetronomeAPIJobDelete = "/v1/jobs/%s"
	// GET /v1/jobs/$jobId
	MetronomeAPIJobGet = "/v1/jobs/%s"
	// GET /v1/jobs
	MetronomeAPIJobList = "/v1/jobs"
	// PUT /v1/jobs/$jobId
	MetronomeAPIJobUpdate = "/v1/jobs/%s"

	// Run model

	// GET /v1/jobs/$jobId/runs
	MetronomeAPIJobRunList = "/v1/jobs/%s/runs/%s"
	// POST /v1/jobs/$jobId/runs
	MetronomeAPIJobRunStart = "/v1/jobs/%s/runs"
	// GET /v1/jobs/$jobId/runs/$runId
	MetronomeAPIJobRunStatus = "/v1/jobs/%s/runs/%s"
	// POST /v1/jobs/$jobId/runs/$runId/action/stop
	MetronomeAPIJobRunStop = "/v1/jobs/%s/runs/%s"

	// Schecule Model

	// POST /v1/jobs/$jobId/schedules
	MetronomeAPIJobScheduleCreate = "/v1/jobs/%s/schedules"
	// GET /v1/jobs/$jobId/schedules/$scheduleId
	MetronomeAPIJobScheduleStatus = "/v1/jobs/%s/schedules/%s"
	// GET /v1/jobs/$jobId/schedules
	MetronomeAPIJobScheduleList = "/v1/jobs/%s/schedules"
	// DELETE /v1/jobs/$jobId/schedules/$scheduleId
	MetronomeAPIJobScheduleDelete = "/v1/jobs/%s/schedules/%s"
	// PUT /v1/jobs/$jobId/schedules/$scheduleId
	MetronomeAPIJobScheduleUpdate = "/v1/jobs/%s/schedules/%s"
	//  GET  /v1/metrics
	MetronomeAPIMetrics = "/v1/metrics"
	//  GET /v1/ping
	MetronomeAPIPing = "/ping"

)
