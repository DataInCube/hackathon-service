package models

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID          string          `json:"id"`
	HackathonID string          `json:"hackathon_id,omitempty"`
	ActorID     string          `json:"actor_id,omitempty"`
	Action      string          `json:"action"`
	Payload     json.RawMessage `json:"payload,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}
