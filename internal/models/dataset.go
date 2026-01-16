package models

import (
	"encoding/json"
	"time"
)

type Dataset struct {
	ID             string          `json:"id"`
	HackathonID    string          `json:"hackathon_id"`
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	SourceURLs     []string        `json:"source_urls,omitempty"`
	ResponseSchema json.RawMessage `json:"response_schema,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type DatasetFile struct {
	ID          string    `json:"id"`
	DatasetID   string    `json:"dataset_id"`
	Name        string    `json:"name"`
	FileType    string    `json:"file_type"`
	Description string    `json:"description,omitempty"`
	URL         string    `json:"url"`
	SizeBytes   int64     `json:"size_bytes,omitempty"`
	Checksum    string    `json:"checksum,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DatasetVariable struct {
	ID          string    `json:"id"`
	DatasetID   string    `json:"dataset_id"`
	Name        string    `json:"name"`
	Role        string    `json:"role"`
	DataType    string    `json:"data_type"`
	Description string    `json:"description,omitempty"`
	Unit        string    `json:"unit,omitempty"`
	Category    string    `json:"category,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
