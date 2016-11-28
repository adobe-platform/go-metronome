package metronome

// Constants defining the various metronome endpoints
const (
	MetronomeAPIJob = "/v1/jobs"
	MetronomeAPIJobs = "/v1/jobs"
	MetronomeAPIStartJob = "/v1/jobs/%s/runs"
	MetronomeAPIDeleteJob = "/v1/jobs/%s"

	MetrononeAPIAddScheduledJob = "scheduler/iso8601"
	MetronomeAPIAddDependentJob = "scheduler/dependency"
)
