package models

import (
	"encoding/json"
	"testing"
)

func TestHTTPError_JSONTags(t *testing.T) {
	e := HTTPError{Code: 400, Message: "bad request"}
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out["code"] != float64(400) {
		t.Fatalf("unexpected code: %v", out["code"])
	}
	if out["message"] != "bad request" {
		t.Fatalf("unexpected message: %v", out["message"])
	}
}
