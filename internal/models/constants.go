package models

const (
	HackathonStateDraft            = "draft"
	HackathonStatePublished        = "published"
	HackathonStateWarmup           = "warmup"
	HackathonStateLive             = "live"
	HackathonStateSubmissionFrozen = "submission_frozen"
	HackathonStateEvaluationOnly   = "evaluation_only"
	HackathonStateCompleted        = "completed"
	HackathonStateArchived         = "archived"
)

const (
	RuleStatusDraft  = "draft"
	RuleStatusLocked = "locked"
)

const (
	SubmissionStatusCreated            = "created"
	SubmissionStatusQueuedForEval      = "queued_for_evaluation"
	SubmissionStatusEvaluationRunning  = "evaluation_running"
	SubmissionStatusEvaluationFailed   = "evaluation_failed"
	SubmissionStatusScored             = "scored"
	SubmissionStatusInvalidated        = "invalidated"
)
