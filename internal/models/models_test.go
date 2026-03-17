package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestConstants_NotEmptyAndDistinct(t *testing.T) {
	states := []string{
		HackathonStateDraft,
		HackathonStatePublished,
		HackathonStateWarmup,
		HackathonStateLive,
		HackathonStateSubmissionFrozen,
		HackathonStateEvaluationOnly,
		HackathonStateCompleted,
		HackathonStateArchived,
	}
	seen := map[string]struct{}{}
	for _, s := range states {
		if s == "" {
			t.Fatal("state constant must not be empty")
		}
		if _, ok := seen[s]; ok {
			t.Fatalf("duplicate state constant %q", s)
		}
		seen[s] = struct{}{}
	}

	if MetricDirectionMinimize == MetricDirectionMaximize {
		t.Fatal("metric directions must be distinct")
	}
	if RuleStatusDraft == RuleStatusLocked {
		t.Fatal("rule statuses must be distinct")
	}
}

func TestModelJSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	activeRuleVersionID := "rulev-1"
	h := Hackathon{
		ID:                  "hack-1",
		Title:               "Milestone Hackathon",
		Description:         "desc",
		State:               HackathonStatePublished,
		Visibility:          "public",
		AllowsTeams:         true,
		RequiresTeams:       true,
		MinTeamSize:         2,
		MaxTeamSize:         5,
		ActiveRuleVersionID: &activeRuleVersionID,
		Metadata:            json.RawMessage(`{"region":"global"}`),
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	b, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("marshal hackathon: %v", err)
	}

	var decoded Hackathon
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal hackathon: %v", err)
	}
	if decoded.ID != h.ID || decoded.State != h.State {
		t.Fatalf("decoded mismatch: %+v", decoded)
	}

	d := Dataset{
		ID:          "data-1",
		HackathonID: "hack-1",
		Title:       "Dataset",
		SourceURLs:  []string{"https://example.com/train.csv"},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if _, err := json.Marshal(d); err != nil {
		t.Fatalf("marshal dataset: %v", err)
	}

	metric := EvaluationMetric{
		ID:          "metric-1",
		HackathonID: "hack-1",
		Name:        "AUC",
		Direction:   MetricDirectionMaximize,
		Scope:       MetricScopeOverall,
		Weight:      1,
		IsPrimary:   true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if _, err := json.Marshal(metric); err != nil {
		t.Fatalf("marshal metric: %v", err)
	}
}
