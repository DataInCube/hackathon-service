package models

import "time"

type Appeal struct {
	ID          string    `json:"id"`
	SubmissionID string   `json:"submission_id"`
	AppellantID string    `json:"appellant_id,omitempty"`
	Content     string    `json:"content"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}
