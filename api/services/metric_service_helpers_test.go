package services

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/DataInCube/hackathon-service/internal/models"
)

func TestMetricNormalizationAndAllowLists(t *testing.T) {
	if normalizeMetricType(" RMSE ") != "rmse" {
		t.Fatalf("normalizeMetricType should lowercase + trim")
	}
	if normalizeDirection(" MAXIMIZE ") != "maximize" {
		t.Fatalf("normalizeDirection should lowercase + trim")
	}
	if normalizeScope(" PER_TARGET ") != "per_target" {
		t.Fatalf("normalizeScope should lowercase + trim")
	}

	if !isAllowedDirection(models.MetricDirectionMinimize) {
		t.Fatalf("minimize should be allowed direction")
	}
	if isAllowedDirection("upwards") {
		t.Fatalf("unexpected allowed direction")
	}

	if !isAllowedScope(models.MetricScopeOverall) {
		t.Fatalf("overall should be allowed scope")
	}
	if isAllowedScope("grouped") {
		t.Fatalf("unexpected allowed scope")
	}

	if !isAllowedMetricType("accuracy") {
		t.Fatalf("accuracy should be allowed metric type")
	}
	if isAllowedMetricType("unknown_metric") {
		t.Fatalf("unknown metric should be rejected")
	}

	if v := nullableString(""); v != nil {
		t.Fatalf("expected nil for empty string, got=%v", v)
	}
	if v := nullableString("target"); v != "target" {
		t.Fatalf("expected original non-empty string, got=%v", v)
	}
}

func TestBuildMetricDefaults(t *testing.T) {
	metric, err := buildMetric("hack_1", models.EvaluationMetric{
		Name:       "  RMSE ",
		MetricType: " RMSE ",
		Direction:  " MINIMIZE ",
		Params:     json.RawMessage(`{"alpha":0.5}`),
	})
	if err != nil {
		t.Fatalf("unexpected buildMetric error: %v", err)
	}
	if metric.Name != "RMSE" {
		t.Fatalf("name should be trimmed, got=%q", metric.Name)
	}
	if metric.Scope != models.MetricScopeOverall {
		t.Fatalf("default scope should be overall, got=%q", metric.Scope)
	}
	if metric.Weight != 1 {
		t.Fatalf("default weight should be 1, got=%v", metric.Weight)
	}
	if metric.HackathonID != "hack_1" {
		t.Fatalf("hackathon id mismatch: %q", metric.HackathonID)
	}
}

func TestBuildMetricTargetScopeInference(t *testing.T) {
	metric, err := buildMetric("hack_2", models.EvaluationMetric{
		Name:           "F1 per label",
		MetricType:     "f1",
		Direction:      "maximize",
		TargetVariable: "label",
	})
	if err != nil {
		t.Fatalf("unexpected buildMetric error: %v", err)
	}
	if metric.Scope != models.MetricScopePerTarget {
		t.Fatalf("expected per_target scope when target_variable is set, got=%q", metric.Scope)
	}
}

func TestBuildMetricValidationFailures(t *testing.T) {
	cases := []models.EvaluationMetric{
		{Name: "", MetricType: "rmse", Direction: "minimize"},
		{Name: "x", MetricType: "nope", Direction: "maximize"},
		{Name: "x", MetricType: "rmse", Direction: "wrong"},
		{Name: "x", MetricType: "rmse", Direction: "minimize", Scope: "per_target"},
		{Name: "x", MetricType: "rmse", Direction: "minimize", Scope: "overall", TargetVariable: "target"},
		{Name: "x", MetricType: "rmse", Direction: "minimize", Weight: -1},
		{Name: "x", MetricType: "rmse", Direction: "minimize", Params: json.RawMessage(`{bad`)},
	}
	for i, input := range cases {
		_, err := buildMetric("hack_bad", input)
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("case %d: expected ErrInvalid, got=%v", i, err)
		}
	}
}
