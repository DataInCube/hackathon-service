package e2e

import "testing"

func TestHackathonServiceCollectionE2E(t *testing.T) {
	runPostmanCollectionE2E(t, postmanRunConfig{
		ServiceName:    "hackathon-service",
		EnvPrefix:      "HACKATHON_E2E_",
		CollectionPath: "../docs/hackathon-service.postman_collection.json",
		RequiredVars: []string{
			"BASE_SERVICE",
			"TOKEN_ADMIN",
			"TOKEN_USER",
			"TOKEN_EVAL",
			"TEAM_ID",
		},
		MinExecuted: 30,
		IDCaptureHints: map[string]string{
			"Create Hackathon":         "HACKATHON_ID",
			"Create Track":             "TRACK_ID",
			"Create Rule":              "RULE_ID",
			"Create Rule Version":      "RULE_VERSION_ID",
			"Create Submission":        "SUBMISSION_ID",
			"Create Resource":          "RESOURCE_ID",
			"Add Dataset File":         "DATA_FILE_ID",
			"Add Dataset Variable":     "VARIABLE_ID",
			"Add Evaluation Metric":    "METRIC_ID",
			"Create Submission Limits": "SUBMISSION_LIMIT_ID",
		},
	})
}
