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

type RuleService struct {
	DB *sql.DB
}

func NewRuleService(db *sql.DB) *RuleService {
	return &RuleService{DB: db}
}

func (s *RuleService) CreateRule(ctx context.Context, hackathonID string, input models.Rule, content json.RawMessage, actorID string) (*models.Rule, *models.RuleVersion, error) {
	if input.Name == "" {
		return nil, nil, fmt.Errorf("rule name is required: %w", ErrInvalid)
	}
	if len(content) == 0 {
		content = json.RawMessage(`{}`)
	}
	if input.TrackID != nil && *input.TrackID != "" {
		var count int
		err := s.DB.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM tracks WHERE id = $1 AND hackathon_id = $2`, *input.TrackID, hackathonID).Scan(&count)
		if err != nil {
			return nil, nil, mapSQLError(err)
		}
		if count == 0 {
			return nil, nil, fmt.Errorf("track_id not found for hackathon: %w", ErrInvalid)
		}
	}

	now := time.Now().UTC()
	rule := models.Rule{
		ID:          uuid.NewString(),
		HackathonID: hackathonID,
		TrackID:     input.TrackID,
		Name:        input.Name,
		Description: input.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	version := models.RuleVersion{
		ID:        uuid.NewString(),
		RuleID:    rule.ID,
		Version:   1,
		Status:    models.RuleStatusDraft,
		Content:   content,
		CreatedBy: actorID,
		CreatedAt: now,
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO rules (id, hackathon_id, track_id, name, description, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		rule.ID, rule.HackathonID, rule.TrackID, rule.Name, rule.Description, rule.CreatedAt, rule.UpdatedAt,
	); err != nil {
		return nil, nil, mapSQLError(err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO rule_versions (id, rule_id, version, status, content, created_by, created_at, locked_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		version.ID, version.RuleID, version.Version, version.Status, version.Content, version.CreatedBy, version.CreatedAt, version.LockedAt,
	); err != nil {
		return nil, nil, mapSQLError(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	return &rule, &version, nil
}

func (s *RuleService) GetByID(ctx context.Context, ruleID string) (*models.Rule, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, hackathon_id, track_id, name, description, created_at, updated_at
		FROM rules WHERE id = $1`, ruleID)

	var rule models.Rule
	if err := row.Scan(&rule.ID, &rule.HackathonID, &rule.TrackID, &rule.Name, &rule.Description, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, mapSQLError(err)
	}
	return &rule, nil
}

func (s *RuleService) GetVersionByID(ctx context.Context, versionID string) (*models.RuleVersion, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, rule_id, version, status, content, created_by, created_at, locked_at
		FROM rule_versions WHERE id = $1`, versionID)

	var v models.RuleVersion
	var content []byte
	if err := row.Scan(&v.ID, &v.RuleID, &v.Version, &v.Status, &content, &v.CreatedBy, &v.CreatedAt, &v.LockedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, mapSQLError(err)
	}
	v.Content = content
	return &v, nil
}

func (s *RuleService) LockVersion(ctx context.Context, versionID string) (*models.RuleVersion, error) {
	v, err := s.GetVersionByID(ctx, versionID)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, fmt.Errorf("rule version not found: %w", ErrNotFound)
	}
	if v.Status == models.RuleStatusLocked {
		return nil, fmt.Errorf("rule version already locked: %w", ErrConflict)
	}

	now := time.Now().UTC()
	_, err = s.DB.ExecContext(ctx, `
		UPDATE rule_versions
		SET status = $1, locked_at = $2
		WHERE id = $3`, models.RuleStatusLocked, now, versionID)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return s.GetVersionByID(ctx, versionID)
}

func (s *RuleService) ListByHackathon(ctx context.Context, hackathonID string, limit, offset int) ([]models.Rule, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, hackathon_id, track_id, name, description, created_at, updated_at
		FROM rules
		WHERE hackathon_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, hackathonID, limit, offset)
	if err != nil {
		return nil, mapSQLError(err)
	}
	defer rows.Close()

	var items []models.Rule
	for rows.Next() {
		var r models.Rule
		if err := rows.Scan(&r.ID, &r.HackathonID, &r.TrackID, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, mapSQLError(err)
		}
		items = append(items, r)
	}
	return items, nil
}

type RuleUpdateInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	TrackID     *string `json:"track_id,omitempty"`
}

func (s *RuleService) Update(ctx context.Context, ruleID string, input RuleUpdateInput) (*models.Rule, error) {
	rule, err := s.GetByID(ctx, ruleID)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, fmt.Errorf("rule not found: %w", ErrNotFound)
	}

	name := rule.Name
	if input.Name != nil {
		if *input.Name == "" {
			return nil, fmt.Errorf("rule name is required: %w", ErrInvalid)
		}
		name = *input.Name
	}
	description := rule.Description
	if input.Description != nil {
		description = *input.Description
	}

	var trackID *string
	if input.TrackID != nil {
		if *input.TrackID == "" {
			trackID = nil
		} else {
			trackID = input.TrackID
		}
	} else {
		trackID = rule.TrackID
	}

	if trackID != nil && *trackID != "" {
		var count int
		err := s.DB.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM tracks WHERE id = $1 AND hackathon_id = $2`, *trackID, rule.HackathonID).Scan(&count)
		if err != nil {
			return nil, mapSQLError(err)
		}
		if count == 0 {
			return nil, fmt.Errorf("track_id not found for hackathon: %w", ErrInvalid)
		}
	}

	_, err = s.DB.ExecContext(ctx, `
		UPDATE rules
		SET name = $1, description = $2, track_id = $3, updated_at = NOW()
		WHERE id = $4`,
		name, description, trackID, ruleID,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return s.GetByID(ctx, ruleID)
}

func (s *RuleService) Delete(ctx context.Context, ruleID string) error {
	var count int
	if err := s.DB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM submissions s
		JOIN rule_versions rv ON s.rule_version_id = rv.id
		WHERE rv.rule_id = $1`, ruleID).Scan(&count); err != nil {
		return mapSQLError(err)
	}
	if count > 0 {
		return fmt.Errorf("rule has submissions and cannot be deleted: %w", ErrConflict)
	}

	if err := s.DB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM hackathons h
		JOIN rule_versions rv ON h.active_rule_version_id = rv.id
		WHERE rv.rule_id = $1`, ruleID).Scan(&count); err != nil {
		return mapSQLError(err)
	}
	if count > 0 {
		return fmt.Errorf("rule is active and cannot be deleted: %w", ErrConflict)
	}

	res, err := s.DB.ExecContext(ctx, `DELETE FROM rules WHERE id = $1`, ruleID)
	if err != nil {
		return mapSQLError(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("rule not found: %w", ErrNotFound)
	}
	return nil
}

func (s *RuleService) CreateVersion(ctx context.Context, ruleID string, content json.RawMessage, actorID string) (*models.RuleVersion, error) {
	if len(content) == 0 {
		content = json.RawMessage(`{}`)
	}

	var versionNum int
	err := s.DB.QueryRowContext(ctx, `SELECT COALESCE(MAX(version),0) FROM rule_versions WHERE rule_id = $1`, ruleID).Scan(&versionNum)
	if err != nil {
		return nil, mapSQLError(err)
	}
	versionNum++

	now := time.Now().UTC()
	version := models.RuleVersion{
		ID:        uuid.NewString(),
		RuleID:    ruleID,
		Version:   versionNum,
		Status:    models.RuleStatusDraft,
		Content:   content,
		CreatedBy: actorID,
		CreatedAt: now,
	}

	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO rule_versions (id, rule_id, version, status, content, created_by, created_at, locked_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		version.ID, version.RuleID, version.Version, version.Status, version.Content, version.CreatedBy, version.CreatedAt, version.LockedAt,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return &version, nil
}

func (s *RuleService) History(ctx context.Context, ruleID string, limit, offset int) ([]models.RuleVersion, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, rule_id, version, status, content, created_by, created_at, locked_at
		FROM rule_versions
		WHERE rule_id = $1
		ORDER BY version DESC
		LIMIT $2 OFFSET $3`, ruleID, limit, offset)
	if err != nil {
		return nil, mapSQLError(err)
	}
	defer rows.Close()

	var items []models.RuleVersion
	for rows.Next() {
		var v models.RuleVersion
		var content []byte
		if err := rows.Scan(&v.ID, &v.RuleID, &v.Version, &v.Status, &content, &v.CreatedBy, &v.CreatedAt, &v.LockedAt); err != nil {
			return nil, mapSQLError(err)
		}
		v.Content = content
		items = append(items, v)
	}
	return items, nil
}

func (s *RuleService) RuleVersionBelongsToHackathon(ctx context.Context, ruleVersionID, hackathonID string) (bool, error) {
	var count int
	err := s.DB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM rule_versions rv
		JOIN rules r ON rv.rule_id = r.id
		WHERE rv.id = $1 AND r.hackathon_id = $2`, ruleVersionID, hackathonID).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, mapSQLError(err)
	}
	return count > 0, nil
}
