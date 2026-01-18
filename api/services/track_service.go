package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/google/uuid"
)

type TrackService struct {
	DB *sql.DB
}

func NewTrackService(db *sql.DB) *TrackService {
	return &TrackService{DB: db}
}

func (s *TrackService) Create(ctx context.Context, hackathonID string, input models.Track) (*models.Track, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("track name is required: %w", ErrInvalid)
	}

	now := time.Now().UTC()
	track := models.Track{
		ID:          uuid.NewString(),
		HackathonID: hackathonID,
		Name:        input.Name,
		Description: input.Description,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO tracks (id, hackathon_id, name, description, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		track.ID, track.HackathonID, track.Name, track.Description, track.IsActive, track.CreatedAt, track.UpdatedAt,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return &track, nil
}

func (s *TrackService) GetByID(ctx context.Context, hackathonID, trackID string) (*models.Track, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, hackathon_id, name, description, is_active, created_at, updated_at
		FROM tracks
		WHERE id = $1 AND hackathon_id = $2`, trackID, hackathonID)

	var t models.Track
	if err := row.Scan(&t.ID, &t.HackathonID, &t.Name, &t.Description, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, mapSQLError(err)
	}
	return &t, nil
}

func (s *TrackService) List(ctx context.Context, hackathonID string, limit, offset int) ([]models.Track, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, hackathon_id, name, description, is_active, created_at, updated_at
		FROM tracks
		WHERE hackathon_id = $1
		ORDER BY created_at
		LIMIT $2 OFFSET $3`, hackathonID, limit, offset)
	if err != nil {
		return nil, mapSQLError(err)
	}
	defer rows.Close()

	var items []models.Track
	for rows.Next() {
		var t models.Track
		if err := rows.Scan(&t.ID, &t.HackathonID, &t.Name, &t.Description, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, mapSQLError(err)
		}
		items = append(items, t)
	}
	return items, nil
}

type TrackUpdateInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

func (s *TrackService) Update(ctx context.Context, hackathonID, trackID string, input TrackUpdateInput) (*models.Track, error) {
	existing, err := s.GetByID(ctx, hackathonID, trackID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("track not found: %w", ErrNotFound)
	}

	name := existing.Name
	if input.Name != nil {
		if *input.Name == "" {
			return nil, fmt.Errorf("track name is required: %w", ErrInvalid)
		}
		name = *input.Name
	}
	description := existing.Description
	if input.Description != nil {
		description = *input.Description
	}
	isActive := existing.IsActive
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	_, err = s.DB.ExecContext(ctx, `
		UPDATE tracks
		SET name = $1, description = $2, is_active = $3, updated_at = NOW()
		WHERE id = $4 AND hackathon_id = $5`,
		name, description, isActive, trackID, hackathonID,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return s.GetByID(ctx, hackathonID, trackID)
}

func (s *TrackService) Delete(ctx context.Context, hackathonID, trackID string) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM tracks WHERE id = $1 AND hackathon_id = $2`, trackID, hackathonID)
	if err != nil {
		return mapSQLError(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("track not found: %w", ErrNotFound)
	}
	return nil
}

func (s *TrackService) Exists(ctx context.Context, hackathonID, trackID string) (bool, error) {
	var count int
	err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM tracks WHERE id = $1 AND hackathon_id = $2`, trackID, hackathonID).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, mapSQLError(err)
	}
	return count > 0, nil
}
