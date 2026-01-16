package models

import (
	"encoding/json"
	"time"
)

type EvaluationMetric struct {
	ID             string          `json:"id"`
	HackathonID    string          `json:"hackathon_id"`
	Name           string          `json:"name"`
	MetricType     string          `json:"metric_type"`
	Direction      string          `json:"direction"`
	Scope          string          `json:"scope"`
	TargetVariable string          `json:"target_variable,omitempty"`
	Weight         float64         `json:"weight"`
	Description    string          `json:"description,omitempty"`
	Params         json.RawMessage `json:"params,omitempty"`
	IsPrimary      bool            `json:"is_primary"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}
