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

type HackathonService struct {
	DB *sql.DB
}

func NewHackathonService(db *sql.DB) *HackathonService {
	return &HackathonService{DB: db}
}

func (s *HackathonService) Create(ctx context.Context, input models.Hackathon, actorID string) (*models.Hackathon, error) {
	if err := validateHackathonInput(input); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	h := input
	h.ID = uuid.NewString()
	h.State = models.HackathonStateDraft
	h.CreatedBy = actorID
	h.CreatedAt = now
	h.UpdatedAt = now
	if h.Visibility == "" {
		h.Visibility = "public"
	}
	if h.Metadata == nil {
		h.Metadata = json.RawMessage(`{}`)
	}

	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO hackathons (
			id, title, description, state, visibility,
			starts_at, ends_at, allows_teams, requires_teams,
			min_team_size, max_team_size, active_rule_version_id,
			leaderboard_frozen, leaderboard_published,
			created_by, metadata, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,
			$6,$7,$8,$9,
			$10,$11,$12,
			$13,$14,
			$15,$16,$17,$18
		)`,
		h.ID, h.Title, h.Description, h.State, h.Visibility,
		h.StartsAt, h.EndsAt, h.AllowsTeams, h.RequiresTeams,
		h.MinTeamSize, h.MaxTeamSize, h.ActiveRuleVersionID,
		h.LeaderboardFrozen, h.LeaderboardPublished,
		h.CreatedBy, h.Metadata, h.CreatedAt, h.UpdatedAt,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}

	return &h, nil
}

func (s *HackathonService) List(ctx context.Context, limit, offset int) ([]models.Hackathon, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, title, description, state, visibility, starts_at, ends_at,
		       allows_teams, requires_teams, min_team_size, max_team_size,
		       active_rule_version_id, leaderboard_frozen, leaderboard_published,
		       created_by, metadata, created_at, updated_at, published_at, completed_at, archived_at
		FROM hackathons
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, mapSQLError(err)
	}
	defer rows.Close()

	var items []models.Hackathon
	for rows.Next() {
		var h models.Hackathon
		var metadata []byte
		if err := rows.Scan(
			&h.ID, &h.Title, &h.Description, &h.State, &h.Visibility, &h.StartsAt, &h.EndsAt,
			&h.AllowsTeams, &h.RequiresTeams, &h.MinTeamSize, &h.MaxTeamSize,
			&h.ActiveRuleVersionID, &h.LeaderboardFrozen, &h.LeaderboardPublished,
			&h.CreatedBy, &metadata, &h.CreatedAt, &h.UpdatedAt, &h.PublishedAt, &h.CompletedAt, &h.ArchivedAt,
		); err != nil {
			return nil, mapSQLError(err)
		}
		h.Metadata = metadata
		items = append(items, h)
	}
	return items, nil
}

func (s *HackathonService) GetByID(ctx context.Context, id string) (*models.Hackathon, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, title, description, state, visibility, starts_at, ends_at,
		       allows_teams, requires_teams, min_team_size, max_team_size,
		       active_rule_version_id, leaderboard_frozen, leaderboard_published,
		       created_by, metadata, created_at, updated_at, published_at, completed_at, archived_at
		FROM hackathons WHERE id = $1`, id)

	var h models.Hackathon
	var metadata []byte
	if err := row.Scan(
		&h.ID, &h.Title, &h.Description, &h.State, &h.Visibility, &h.StartsAt, &h.EndsAt,
		&h.AllowsTeams, &h.RequiresTeams, &h.MinTeamSize, &h.MaxTeamSize,
		&h.ActiveRuleVersionID, &h.LeaderboardFrozen, &h.LeaderboardPublished,
		&h.CreatedBy, &metadata, &h.CreatedAt, &h.UpdatedAt, &h.PublishedAt, &h.CompletedAt, &h.ArchivedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, mapSQLError(err)
	}
	h.Metadata = metadata
	return &h, nil
}

func (s *HackathonService) Update(ctx context.Context, id string, input models.Hackathon) (*models.Hackathon, error) {
	if err := validateHackathonInput(input); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	res, err := s.DB.ExecContext(ctx, `
		UPDATE hackathons
		SET title = $1, description = $2, visibility = $3, starts_at = $4, ends_at = $5,
		    allows_teams = $6, requires_teams = $7, min_team_size = $8, max_team_size = $9,
		    metadata = $10, updated_at = $11
		WHERE id = $12`,
		input.Title, input.Description, input.Visibility, input.StartsAt, input.EndsAt,
		input.AllowsTeams, input.RequiresTeams, input.MinTeamSize, input.MaxTeamSize,
		normalizeMetadata(input.Metadata), now, id,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, fmt.Errorf("hackathon not found: %w", ErrNotFound)
	}

	return s.GetByID(ctx, id)
}

func (s *HackathonService) Delete(ctx context.Context, id string) error {
	h, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if h == nil {
		return fmt.Errorf("hackathon not found: %w", ErrNotFound)
	}
	if h.State != models.HackathonStateDraft {
		return fmt.Errorf("only draft hackathons can be deleted: %w", ErrInvalid)
	}

	res, err := s.DB.ExecContext(ctx, `DELETE FROM hackathons WHERE id = $1`, id)
	if err != nil {
		return mapSQLError(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("hackathon not found: %w", ErrNotFound)
	}
	return nil
}

func (s *HackathonService) Publish(ctx context.Context, id string) (*models.Hackathon, error) {
	return s.transition(ctx, id, models.HackathonStatePublished)
}

func (s *HackathonService) Transition(ctx context.Context, id, target string) (*models.Hackathon, error) {
	return s.transition(ctx, id, target)
}

func (s *HackathonService) GetState(ctx context.Context, id string) (string, error) {
	var state string
	err := s.DB.QueryRowContext(ctx, `SELECT state FROM hackathons WHERE id = $1`, id).Scan(&state)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("hackathon not found: %w", ErrNotFound)
	}
	if err != nil {
		return "", mapSQLError(err)
	}
	return state, nil
}

func (s *HackathonService) SetActiveRuleVersion(ctx context.Context, hackathonID, ruleVersionID string) error {
	res, err := s.DB.ExecContext(ctx, `
		UPDATE hackathons SET active_rule_version_id = $1, updated_at = NOW()
		WHERE id = $2`, ruleVersionID, hackathonID)
	if err != nil {
		return mapSQLError(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("hackathon not found: %w", ErrNotFound)
	}
	return nil
}

func (s *HackathonService) GetTeamPolicy(ctx context.Context, id string) (*models.TeamPolicy, error) {
	var policy models.TeamPolicy
	err := s.DB.QueryRowContext(ctx, `
		SELECT id, allows_teams, requires_teams, min_team_size, max_team_size
		FROM hackathons WHERE id = $1`, id).
		Scan(&policy.HackathonID, &policy.AllowsTeams, &policy.RequiresTeams, &policy.MinTeamSize, &policy.MaxTeamSize)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("hackathon not found: %w", ErrNotFound)
	}
	if err != nil {
		return nil, mapSQLError(err)
	}
	return &policy, nil
}

func (s *HackathonService) GetLeaderboardPolicy(ctx context.Context, id string) (*models.LeaderboardPolicy, error) {
	var policy models.LeaderboardPolicy
	err := s.DB.QueryRowContext(ctx, `
		SELECT id, leaderboard_frozen, leaderboard_published
		FROM hackathons WHERE id = $1`, id).
		Scan(&policy.HackathonID, &policy.Frozen, &policy.Published)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("hackathon not found: %w", ErrNotFound)
	}
	if err != nil {
		return nil, mapSQLError(err)
	}
	return &policy, nil
}

func (s *HackathonService) FreezeLeaderboard(ctx context.Context, id string) (*models.Hackathon, error) {
	h, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if h == nil {
		return nil, fmt.Errorf("hackathon not found: %w", ErrNotFound)
	}
	if !isStateAtLeast(h.State, models.HackathonStateSubmissionFrozen) {
		return nil, fmt.Errorf("hackathon must be submission_frozen or later to freeze leaderboard: %w", ErrInvalid)
	}

	res, err := s.DB.ExecContext(ctx, `
		UPDATE hackathons SET leaderboard_frozen = true, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return nil, mapSQLError(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, fmt.Errorf("hackathon not found: %w", ErrNotFound)
	}
	return s.GetByID(ctx, id)
}

func (s *HackathonService) PublishLeaderboard(ctx context.Context, id string) (*models.Hackathon, error) {
	h, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if h == nil {
		return nil, fmt.Errorf("hackathon not found: %w", ErrNotFound)
	}
	if !isStateAtLeast(h.State, models.HackathonStateCompleted) {
		return nil, fmt.Errorf("hackathon must be completed to publish leaderboard: %w", ErrInvalid)
	}

	res, err := s.DB.ExecContext(ctx, `
		UPDATE hackathons SET leaderboard_published = true, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return nil, mapSQLError(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, fmt.Errorf("hackathon not found: %w", ErrNotFound)
	}
	return s.GetByID(ctx, id)
}

func (s *HackathonService) UnfreezeLeaderboard(ctx context.Context, id string) (*models.Hackathon, error) {
	h, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if h == nil {
		return nil, fmt.Errorf("hackathon not found: %w", ErrNotFound)
	}
	if !isStateAtLeast(h.State, models.HackathonStateSubmissionFrozen) {
		return nil, fmt.Errorf("hackathon must be submission_frozen or later to unfreeze leaderboard: %w", ErrInvalid)
	}

	res, err := s.DB.ExecContext(ctx, `
		UPDATE hackathons SET leaderboard_frozen = false, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return nil, mapSQLError(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, fmt.Errorf("hackathon not found: %w", ErrNotFound)
	}
	return s.GetByID(ctx, id)
}

func (s *HackathonService) transition(ctx context.Context, id, target string) (*models.Hackathon, error) {
	h, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if h == nil {
		return nil, fmt.Errorf("hackathon not found: %w", ErrNotFound)
	}

	if h.State == target {
		return nil, fmt.Errorf("hackathon already in state %s: %w", target, ErrConflict)
	}

	if !isTransitionAllowed(h.State, target) {
		return nil, fmt.Errorf("transition not allowed from %s to %s: %w", h.State, target, ErrInvalid)
	}

	if target == models.HackathonStateLive && h.ActiveRuleVersionID == nil {
		return nil, fmt.Errorf("active rule version required before live: %w", ErrInvalid)
	}

	now := time.Now().UTC()
	var publishedAt, completedAt, archivedAt *time.Time
	if target == models.HackathonStatePublished {
		publishedAt = &now
	}
	if target == models.HackathonStateCompleted {
		completedAt = &now
	}
	if target == models.HackathonStateArchived {
		archivedAt = &now
	}

	_, err = s.DB.ExecContext(ctx, `
		UPDATE hackathons
		SET state = $1, published_at = COALESCE($2, published_at),
		    completed_at = COALESCE($3, completed_at),
		    archived_at = COALESCE($4, archived_at),
		    updated_at = NOW()
		WHERE id = $5`,
		target, publishedAt, completedAt, archivedAt, h.ID,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}

	return s.GetByID(ctx, id)
}

func validateHackathonInput(h models.Hackathon) error {
	if h.Title == "" {
		return fmt.Errorf("title is required: %w", ErrInvalid)
	}
	if h.StartsAt != nil && h.EndsAt != nil && h.EndsAt.Before(*h.StartsAt) {
		return fmt.Errorf("ends_at must be after starts_at: %w", ErrInvalid)
	}
	if h.RequiresTeams && !h.AllowsTeams {
		return fmt.Errorf("requires_teams=true requires allows_teams=true: %w", ErrInvalid)
	}
	if h.MinTeamSize < 0 || h.MaxTeamSize < 0 {
		return fmt.Errorf("team sizes must be >= 0: %w", ErrInvalid)
	}
	if h.MaxTeamSize > 0 && h.MinTeamSize > h.MaxTeamSize {
		return fmt.Errorf("min_team_size cannot exceed max_team_size: %w", ErrInvalid)
	}
	return nil
}

func normalizeMetadata(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(`{}`)
	}
	return raw
}

func isTransitionAllowed(current, target string) bool {
	allowed := map[string][]string{
		models.HackathonStateDraft:            {models.HackathonStatePublished},
		models.HackathonStatePublished:        {models.HackathonStateWarmup},
		models.HackathonStateWarmup:           {models.HackathonStateLive},
		models.HackathonStateLive:             {models.HackathonStateSubmissionFrozen},
		models.HackathonStateSubmissionFrozen: {models.HackathonStateEvaluationOnly},
		models.HackathonStateEvaluationOnly:   {models.HackathonStateCompleted},
		models.HackathonStateCompleted:        {models.HackathonStateArchived},
	}

	for _, next := range allowed[current] {
		if next == target {
			return true
		}
	}
	return false
}

func isStateAtLeast(state, floor string) bool {
	order := map[string]int{
		models.HackathonStateDraft:            0,
		models.HackathonStatePublished:        1,
		models.HackathonStateWarmup:           2,
		models.HackathonStateLive:             3,
		models.HackathonStateSubmissionFrozen: 4,
		models.HackathonStateEvaluationOnly:   5,
		models.HackathonStateCompleted:        6,
		models.HackathonStateArchived:         7,
	}
	return order[state] >= order[floor]
}
