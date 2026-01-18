package services

import (
	"encoding/json"
	"testing"
)

func TestIsSubmissionTransitionAllowed(t *testing.T) {
	if !isSubmissionTransitionAllowed("queued_for_evaluation", "evaluation_running") {
		t.Fatalf("expected queued_for_evaluation -> evaluation_running allowed")
	}
	if !isSubmissionTransitionAllowed("evaluation_running", "scored") {
		t.Fatalf("expected evaluation_running -> scored allowed")
	}
	if isSubmissionTransitionAllowed("created", "evaluation_running") {
		t.Fatalf("expected created -> evaluation_running disallowed")
	}
}

func TestMergeMetadata(t *testing.T) {
	existing := json.RawMessage(`{"a":1,"b":"x"}`)
	patch := json.RawMessage(`{"b":"y","c":true}`)

	merged, err := mergeMetadata(existing, patch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(merged, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if out["a"].(float64) != 1 {
		t.Fatalf("expected a=1")
	}
	if out["b"].(string) != "y" {
		t.Fatalf("expected b=y")
	}
	if out["c"].(bool) != true {
		t.Fatalf("expected c=true")
	}

	_, err = mergeMetadata(existing, json.RawMessage(`{bad json}`))
	if err == nil {
		t.Fatalf("expected error on invalid json")
	}
}
