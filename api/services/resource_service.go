package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/google/uuid"
)

type ResourceService struct {
	DB *sql.DB
}

func NewResourceService(db *sql.DB) *ResourceService {
	return &ResourceService{DB: db}
}

func (s *ResourceService) Create(ctx context.Context, hackathonID string, input models.Resource) (*models.Resource, error) {
	if input.Title == "" || input.URL == "" {
		return nil, fmt.Errorf("resource title and url are required: %w", ErrInvalid)
	}
	if input.Type == "" {
		input.Type = "resource"
	}

	now := time.Now().UTC()
	res := models.Resource{
		ID:          uuid.NewString(),
		HackathonID: hackathonID,
		Type:        input.Type,
		Title:       input.Title,
		URL:         input.URL,
		Metadata:    normalizeMetadata(input.Metadata),
		CreatedAt:   now,
	}

	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO resources (id, hackathon_id, type, title, url, metadata, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		res.ID, res.HackathonID, res.Type, res.Title, res.URL, res.Metadata, res.CreatedAt,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return &res, nil
}

func (s *ResourceService) List(ctx context.Context, hackathonID string, limit, offset int) ([]models.Resource, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, hackathon_id, type, title, url, metadata, created_at
		FROM resources
		WHERE hackathon_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, hackathonID, limit, offset)
	if err != nil {
		return nil, mapSQLError(err)
	}
	defer rows.Close()

	var items []models.Resource
	for rows.Next() {
		var r models.Resource
		var metadata []byte
		if err := rows.Scan(&r.ID, &r.HackathonID, &r.Type, &r.Title, &r.URL, &metadata, &r.CreatedAt); err != nil {
			return nil, mapSQLError(err)
		}
		r.Metadata = metadata
		items = append(items, r)
	}
	return items, nil
}

func (s *ResourceService) GetByID(ctx context.Context, hackathonID, resourceID string) (*models.Resource, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, hackathon_id, type, title, url, metadata, created_at
		FROM resources
		WHERE id = $1 AND hackathon_id = $2`, resourceID, hackathonID)
	var r models.Resource
	var metadata []byte
	if err := row.Scan(&r.ID, &r.HackathonID, &r.Type, &r.Title, &r.URL, &metadata, &r.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, mapSQLError(err)
	}
	r.Metadata = metadata
	return &r, nil
}

type ResourceUpdateInput struct {
	Type     *string          `json:"type,omitempty"`
	Title    *string          `json:"title,omitempty"`
	URL      *string          `json:"url,omitempty"`
	Metadata *json.RawMessage `json:"metadata,omitempty"`
}

func (s *ResourceService) Update(ctx context.Context, hackathonID, resourceID string, input ResourceUpdateInput) (*models.Resource, error) {
	existing, err := s.GetByID(ctx, hackathonID, resourceID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("resource not found: %w", ErrNotFound)
	}

	resourceType := existing.Type
	if input.Type != nil {
		if *input.Type == "" {
			return nil, fmt.Errorf("resource type is required: %w", ErrInvalid)
		}
		resourceType = *input.Type
	}
	title := existing.Title
	if input.Title != nil {
		if *input.Title == "" {
			return nil, fmt.Errorf("resource title is required: %w", ErrInvalid)
		}
		title = *input.Title
	}
	url := existing.URL
	if input.URL != nil {
		if *input.URL == "" {
			return nil, fmt.Errorf("resource url is required: %w", ErrInvalid)
		}
		url = *input.URL
	}
	metadata := existing.Metadata
	if input.Metadata != nil {
		metadata = normalizeMetadata(*input.Metadata)
	}

	_, err = s.DB.ExecContext(ctx, `
		UPDATE resources
		SET type = $1, title = $2, url = $3, metadata = $4
		WHERE id = $5 AND hackathon_id = $6`,
		resourceType, title, url, metadata, resourceID, hackathonID,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return s.GetByID(ctx, hackathonID, resourceID)
}

func (s *ResourceService) Delete(ctx context.Context, hackathonID, resourceID string) error {
	res, err := s.DB.ExecContext(ctx, `
		DELETE FROM resources WHERE id = $1 AND hackathon_id = $2`, resourceID, hackathonID)
	if err != nil {
		return mapSQLError(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("resource not found: %w", ErrNotFound)
	}
	return nil
}
