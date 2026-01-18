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
	SubmissionStatusCreated           = "created"
	SubmissionStatusQueuedForEval     = "queued_for_evaluation"
	SubmissionStatusEvaluationRunning = "evaluation_running"
	SubmissionStatusEvaluationFailed  = "evaluation_failed"
	SubmissionStatusScored            = "scored"
	SubmissionStatusInvalidated       = "invalidated"
)

const (
	DatasetFileTypeTrain            = "train"
	DatasetFileTypeTest             = "test"
	DatasetFileTypeSampleSubmission = "sample_submission"
	DatasetFileTypeDataDictionary   = "data_dictionary"
	DatasetFileTypeOther            = "other"
)

const (
	DatasetVariableRoleFeature    = "feature"
	DatasetVariableRoleTarget     = "target"
	DatasetVariableRoleIdentifier = "identifier"
	DatasetVariableRoleMetadata   = "metadata"
)

const (
	DatasetDataTypeString      = "string"
	DatasetDataTypeInteger     = "integer"
	DatasetDataTypeFloat       = "float"
	DatasetDataTypeBoolean     = "boolean"
	DatasetDataTypeCategorical = "categorical"
	DatasetDataTypeDatetime    = "datetime"
	DatasetDataTypeText        = "text"
)

const (
	MetricDirectionMinimize = "minimize"
	MetricDirectionMaximize = "maximize"
)

const (
	MetricScopeOverall   = "overall"
	MetricScopePerTarget = "per_target"
)
