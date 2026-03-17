package main

import (
	"testing"
)

func TestEventSubjectsFromEnv_Defaults(t *testing.T) {
	subjects := eventSubjectsFromEnv(func(_ string, def string) string { return def })
	if len(subjects) < 25 {
		t.Fatalf("expected many default subjects, got %d", len(subjects))
	}

	required := map[string]bool{
		"hackathon.created":              false,
		"hackathon.phase.changed":        false,
		"submission.created":             false,
		"evaluation.completed":           false,
		"leaderboard.freeze.requested":   false,
		"leaderboard.unfreeze.requested": false,
	}
	for _, s := range subjects {
		if _, ok := required[s]; ok {
			required[s] = true
		}
	}
	for subject, found := range required {
		if !found {
			t.Fatalf("expected default subject %q in %v", subject, subjects)
		}
	}
}

func TestEventSubjectsFromEnv_Overrides(t *testing.T) {
	subjects := eventSubjectsFromEnv(func(key, def string) string {
		if key == "NATS_SUBJECT_HACKATHON_CREATED" {
			return "custom.hackathon.created"
		}
		return def
	})
	if subjects[0] != "custom.hackathon.created" {
		t.Fatalf("expected first subject override, got %q", subjects[0])
	}
}
