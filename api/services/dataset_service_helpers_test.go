package services

import (
	"errors"
	"reflect"
	"testing"

	"github.com/DataInCube/hackathon-service/internal/models"
)

func TestNormalizeURLList(t *testing.T) {
	got, err := normalizeURLList([]string{" https://a.example ", "https://b.example"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"https://a.example", "https://b.example"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected normalized urls: got=%v want=%v", got, want)
	}

	got, err = normalizeURLList(nil)
	if err != nil {
		t.Fatalf("unexpected error for empty input: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice for empty input, got=%v", got)
	}

	_, err = normalizeURLList([]string{"ok", "  "})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid for empty url item, got=%v", err)
	}
}

func TestDatasetEnumHelpers(t *testing.T) {
	if normalizeFileType(" TRAIN ") != "train" {
		t.Fatalf("normalizeFileType should lowercase + trim")
	}
	if !isAllowedFileType(models.DatasetFileTypeSampleSubmission) {
		t.Fatalf("expected sample_submission file type to be allowed")
	}
	if isAllowedFileType("binary_dump") {
		t.Fatalf("expected unknown file type to be rejected")
	}

	if normalizeRole(" TARGET ") != "target" {
		t.Fatalf("normalizeRole should lowercase + trim")
	}
	if !isAllowedRole(models.DatasetVariableRoleMetadata) {
		t.Fatalf("expected metadata role to be allowed")
	}
	if isAllowedRole("label") {
		t.Fatalf("expected unknown role to be rejected")
	}

	if normalizeDataType(" FLOAT ") != "float" {
		t.Fatalf("normalizeDataType should lowercase + trim")
	}
	if !isAllowedDataType(models.DatasetDataTypeDatetime) {
		t.Fatalf("expected datetime type to be allowed")
	}
	if isAllowedDataType("decimal128") {
		t.Fatalf("expected unknown data type to be rejected")
	}
}
