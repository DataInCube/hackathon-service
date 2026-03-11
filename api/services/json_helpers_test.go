package services

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestEnsureValidJSON(t *testing.T) {
	if err := ensureValidJSON(nil, "payload"); err != nil {
		t.Fatalf("nil json should be allowed: %v", err)
	}

	if err := ensureValidJSON(json.RawMessage(`{"ok":true}`), "payload"); err != nil {
		t.Fatalf("valid json should pass: %v", err)
	}

	err := ensureValidJSON(json.RawMessage(`{broken`), "payload")
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid for malformed json, got=%v", err)
	}
}
