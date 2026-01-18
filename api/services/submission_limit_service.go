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

type SubmissionLimitService struct {
	DB *sql.DB
}

func NewSubmissionLimitService(db *sql.DB) *SubmissionLimitService {
	return &SubmissionLimitService{DB: db}
}

func (s *SubmissionLimitService) Create(ctx context.Context, hackathonID string, input models.SubmissionLimit) (*models.SubmissionLimit, error) {
	if err := ensureEditableHackathon(ctx, s.DB, hackathonID); err != nil {
		return nil, err
	}
	if err := validateSubmissionLimits(input.PerDay, input.Total, input.PerTeam); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	limit := models.SubmissionLimit{
		ID:          uuid.NewString(),
		HackathonID: hackathonID,
		PerDay:      input.PerDay,
		Total:       input.Total,
		PerTeam:     input.PerTeam,
		Notes:       input.Notes,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO submission_limits (id, hackathon_id, per_day, total, per_team, notes, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		limit.ID, limit.HackathonID, limit.PerDay, limit.Total, limit.PerTeam, limit.Notes, limit.CreatedAt, limit.UpdatedAt,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return &limit, nil
}

func (s *SubmissionLimitService) Get(ctx context.Context, hackathonID string) (*models.SubmissionLimit, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, hackathon_id, per_day, total, per_team, notes, created_at, updated_at
		FROM submission_limits
		WHERE hackathon_id = $1`, hackathonID)

	var limit models.SubmissionLimit
	if err := row.Scan(&limit.ID, &limit.HackathonID, &limit.PerDay, &limit.Total, &limit.PerTeam, &limit.Notes, &limit.CreatedAt, &limit.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, mapSQLError(err)
	}
	return &limit, nil
}

type SubmissionLimitUpdateInput struct {
	PerDay  *int    `json:"per_day,omitempty"`
	Total   *int    `json:"total,omitempty"`
	PerTeam *int    `json:"per_team,omitempty"`
	Notes   *string `json:"notes,omitempty"`
}

func (s *SubmissionLimitService) Update(ctx context.Context, hackathonID string, input SubmissionLimitUpdateInput) (*models.SubmissionLimit, error) {
	if err := ensureEditableHackathon(ctx, s.DB, hackathonID); err != nil {
		return nil, err
	}
	existing, err := s.Get(ctx, hackathonID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("submission limits not found: %w", ErrNotFound)
	}

	perDay := existing.PerDay
	if input.PerDay != nil {
		perDay = *input.PerDay
	}
	total := existing.Total
	if input.Total != nil {
		total = *input.Total
	}
	perTeam := existing.PerTeam
	if input.PerTeam != nil {
		perTeam = *input.PerTeam
	}
	if err := validateSubmissionLimits(perDay, total, perTeam); err != nil {
		return nil, err
	}
	notes := existing.Notes
	if input.Notes != nil {
		notes = *input.Notes
	}

	_, err = s.DB.ExecContext(ctx, `
		UPDATE submission_limits
		SET per_day = $1, total = $2, per_team = $3, notes = $4, updated_at = NOW()
		WHERE hackathon_id = $5`, perDay, total, perTeam, notes, hackathonID)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return s.Get(ctx, hackathonID)
}

func (s *SubmissionLimitService) Delete(ctx context.Context, hackathonID string) error {
	if err := ensureEditableHackathon(ctx, s.DB, hackathonID); err != nil {
		return err
	}
	res, err := s.DB.ExecContext(ctx, `DELETE FROM submission_limits WHERE hackathon_id = $1`, hackathonID)
	if err != nil {
		return mapSQLError(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("submission limits not found: %w", ErrNotFound)
	}
	return nil
}

func validateSubmissionLimits(perDay, total, perTeam int) error {
	if perDay < 0 || total < 0 || perTeam < 0 {
		return fmt.Errorf("submission limits must be >= 0: %w", ErrInvalid)
	}
	return nil
}
