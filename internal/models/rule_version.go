package models

import (
	"encoding/json"
	"time"
)

type RuleVersion struct {
	ID        string          `json:"id"`
	RuleID    string          `json:"rule_id"`
	Version   int             `json:"version"`
	Status    string          `json:"status"`
	Content   json.RawMessage `json:"content"`
	CreatedBy string          `json:"created_by,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	LockedAt  *time.Time      `json:"locked_at,omitempty"`
}
