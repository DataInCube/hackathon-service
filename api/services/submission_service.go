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

type SubmissionService struct {
	DB          *sql.DB
	TrackLookup *TrackService
}

func NewSubmissionService(db *sql.DB, trackLookup *TrackService) *SubmissionService {
	return &SubmissionService{DB: db, TrackLookup: trackLookup}
}

type SubmissionInput struct {
	TrackID     *string         `json:"track_id,omitempty"`
	TeamID      *string         `json:"team_id,omitempty"`
	MemberCount *int            `json:"member_count,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
}

type SubmissionUpdateInput struct {
	TeamID      *string         `json:"team_id,omitempty"`
	MemberCount *int            `json:"member_count,omitempty"`
	Metadata    *json.RawMessage `json:"metadata,omitempty"`
}

func (s *SubmissionService) Create(ctx context.Context, hackathonID string, input SubmissionInput, actorID string) (*models.Submission, error) {
	state, ruleVersionID, policy, err := s.loadHackathonForSubmission(ctx, hackathonID)
	if err != nil {
		return nil, err
	}
	if state != models.HackathonStateLive {
		return nil, fmt.Errorf("hackathon not live: %w", ErrInvalid)
	}
	if ruleVersionID == "" {
		return nil, fmt.Errorf("active rule version required: %w", ErrInvalid)
	}

	if policy.RequiresTeams && (input.TeamID == nil || *input.TeamID == "") {
		return nil, fmt.Errorf("team_id required: %w", ErrInvalid)
	}
	if !policy.AllowsTeams && input.TeamID != nil && *input.TeamID != "" {
		return nil, fmt.Errorf("teams not allowed: %w", ErrInvalid)
	}
	if input.MemberCount != nil {
		if policy.MinTeamSize > 0 && *input.MemberCount < policy.MinTeamSize {
			return nil, fmt.Errorf("team too small: %w", ErrInvalid)
		}
		if policy.MaxTeamSize > 0 && *input.MemberCount > policy.MaxTeamSize {
			return nil, fmt.Errorf("team too large: %w", ErrInvalid)
		}
	}

	if input.TrackID != nil && *input.TrackID != "" && s.TrackLookup != nil {
		ok, err := s.TrackLookup.Exists(ctx, hackathonID, *input.TrackID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("track_id not found: %w", ErrInvalid)
		}
	}

	now := time.Now().UTC()
	sub := models.Submission{
		ID:            uuid.NewString(),
		HackathonID:   hackathonID,
		TrackID:       input.TrackID,
		RuleVersionID: ruleVersionID,
		SubmittedBy:   actorID,
		TeamID:        input.TeamID,
		Status:        models.SubmissionStatusCreated,
		Phase:         state,
		Metadata:      normalizeMetadata(input.Metadata),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO submissions (
			id, hackathon_id, track_id, rule_version_id, submitted_by,
			team_id, status, phase, metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		sub.ID, sub.HackathonID, sub.TrackID, sub.RuleVersionID, sub.SubmittedBy,
		sub.TeamID, sub.Status, sub.Phase, sub.Metadata, sub.CreatedAt, sub.UpdatedAt,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}

	return &sub, nil
}

func (s *SubmissionService) GetByID(ctx context.Context, id string) (*models.Submission, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, hackathon_id, track_id, rule_version_id, submitted_by,
		       team_id, status, phase, metadata, created_at, updated_at, locked_at, invalidated_at
		FROM submissions WHERE id = $1`, id)

	var sub models.Submission
	var metadata []byte
	if err := row.Scan(
		&sub.ID, &sub.HackathonID, &sub.TrackID, &sub.RuleVersionID, &sub.SubmittedBy,
		&sub.TeamID, &sub.Status, &sub.Phase, &metadata, &sub.CreatedAt, &sub.UpdatedAt, &sub.LockedAt, &sub.InvalidatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, mapSQLError(err)
	}
	sub.Metadata = metadata
	return &sub, nil
}

func (s *SubmissionService) ListByHackathon(ctx context.Context, hackathonID string, limit, offset int) ([]models.Submission, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, hackathon_id, track_id, rule_version_id, submitted_by,
		       team_id, status, phase, metadata, created_at, updated_at, locked_at, invalidated_at
		FROM submissions
		WHERE hackathon_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, hackathonID, limit, offset)
	if err != nil {
		return nil, mapSQLError(err)
	}
	defer rows.Close()

	var items []models.Submission
	for rows.Next() {
		var sub models.Submission
		var metadata []byte
		if err := rows.Scan(
			&sub.ID, &sub.HackathonID, &sub.TrackID, &sub.RuleVersionID, &sub.SubmittedBy,
			&sub.TeamID, &sub.Status, &sub.Phase, &metadata, &sub.CreatedAt, &sub.UpdatedAt, &sub.LockedAt, &sub.InvalidatedAt,
		); err != nil {
			return nil, mapSQLError(err)
		}
		sub.Metadata = metadata
		items = append(items, sub)
	}
	return items, nil
}

func (s *SubmissionService) Update(ctx context.Context, id string, input SubmissionUpdateInput) (*models.Submission, error) {
	sub, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, fmt.Errorf("submission not found: %w", ErrNotFound)
	}
	if sub.Status != models.SubmissionStatusCreated {
		return nil, fmt.Errorf("only created submissions can be updated: %w", ErrInvalid)
	}

	state, _, policy, err := s.loadHackathonForSubmission(ctx, sub.HackathonID)
	if err != nil {
		return nil, err
	}
	if state != models.HackathonStateLive {
		return nil, fmt.Errorf("hackathon not live: %w", ErrInvalid)
	}

	teamID := sub.TeamID
	if input.TeamID != nil {
		if *input.TeamID == "" {
			teamID = nil
		} else {
			teamID = input.TeamID
		}
	}

	if policy.RequiresTeams && (teamID == nil || *teamID == "") {
		return nil, fmt.Errorf("team_id required: %w", ErrInvalid)
	}
	if !policy.AllowsTeams && teamID != nil && *teamID != "" {
		return nil, fmt.Errorf("teams not allowed: %w", ErrInvalid)
	}
	if input.MemberCount != nil {
		if policy.MinTeamSize > 0 && *input.MemberCount < policy.MinTeamSize {
			return nil, fmt.Errorf("team too small: %w", ErrInvalid)
		}
		if policy.MaxTeamSize > 0 && *input.MemberCount > policy.MaxTeamSize {
			return nil, fmt.Errorf("team too large: %w", ErrInvalid)
		}
	}

	metadata := sub.Metadata
	if input.Metadata != nil {
		merged, err := mergeMetadata(sub.Metadata, *input.Metadata)
		if err != nil {
			return nil, fmt.Errorf("invalid metadata: %w", ErrInvalid)
		}
		metadata = merged
	}

	_, err = s.DB.ExecContext(ctx, `
		UPDATE submissions
		SET team_id = $1, metadata = $2, updated_at = NOW()
		WHERE id = $3`,
		teamID, metadata, id,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return s.GetByID(ctx, id)
}

func (s *SubmissionService) Delete(ctx context.Context, id string) error {
	sub, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if sub == nil {
		return fmt.Errorf("submission not found: %w", ErrNotFound)
	}
	if sub.Status != models.SubmissionStatusCreated {
		return fmt.Errorf("only created submissions can be deleted: %w", ErrInvalid)
	}

	res, err := s.DB.ExecContext(ctx, `DELETE FROM submissions WHERE id = $1`, id)
	if err != nil {
		return mapSQLError(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SubmissionService) Lock(ctx context.Context, id string) (*models.Submission, error) {
	sub, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, fmt.Errorf("submission not found: %w", ErrNotFound)
	}
	if sub.Status != models.SubmissionStatusCreated {
		return nil, fmt.Errorf("only created submissions can be locked: %w", ErrInvalid)
	}

	now := time.Now().UTC()
	_, err = s.DB.ExecContext(ctx, `
		UPDATE submissions
		SET status = $1, locked_at = $2, updated_at = NOW()
		WHERE id = $3`,
		models.SubmissionStatusQueuedForEval, now, id,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return s.GetByID(ctx, id)
}

func (s *SubmissionService) UpdateEvaluationStatus(ctx context.Context, id, target string, metadataPatch *json.RawMessage) (*models.Submission, error) {
	sub, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, fmt.Errorf("submission not found: %w", ErrNotFound)
	}
	if !isSubmissionTransitionAllowed(sub.Status, target) {
		return nil, fmt.Errorf("invalid submission status transition: %w", ErrInvalid)
	}

	metadata := sub.Metadata
	if metadataPatch != nil {
		merged, err := mergeMetadata(sub.Metadata, *metadataPatch)
		if err != nil {
			return nil, fmt.Errorf("invalid metadata: %w", ErrInvalid)
		}
		metadata = merged
	}

	_, err = s.DB.ExecContext(ctx, `
		UPDATE submissions
		SET status = $1, metadata = $2, updated_at = NOW()
		WHERE id = $3`, target, metadata, id)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return s.GetByID(ctx, id)
}

func (s *SubmissionService) Invalidate(ctx context.Context, id string) (*models.Submission, error) {
	sub, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, fmt.Errorf("submission not found: %w", ErrNotFound)
	}
	if sub.Status == models.SubmissionStatusInvalidated {
		return nil, fmt.Errorf("submission already invalidated: %w", ErrConflict)
	}

	now := time.Now().UTC()
	_, err = s.DB.ExecContext(ctx, `
		UPDATE submissions
		SET status = $1, invalidated_at = $2, updated_at = NOW()
		WHERE id = $3`,
		models.SubmissionStatusInvalidated, now, id,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return s.GetByID(ctx, id)
}

func isSubmissionTransitionAllowed(current, target string) bool {
	allowed := map[string][]string{
		models.SubmissionStatusQueuedForEval:     {models.SubmissionStatusEvaluationRunning, models.SubmissionStatusEvaluationFailed},
		models.SubmissionStatusEvaluationRunning: {models.SubmissionStatusEvaluationFailed, models.SubmissionStatusScored},
	}
	for _, next := range allowed[current] {
		if next == target {
			return true
		}
	}
	return false
}

func (s *SubmissionService) loadHackathonForSubmission(ctx context.Context, hackathonID string) (string, string, models.TeamPolicy, error) {
	var state string
	var ruleID sql.NullString
	var policy models.TeamPolicy
	err := s.DB.QueryRowContext(ctx, `
		SELECT state, active_rule_version_id, allows_teams, requires_teams, min_team_size, max_team_size
		FROM hackathons WHERE id = $1`, hackathonID).
		Scan(&state, &ruleID, &policy.AllowsTeams, &policy.RequiresTeams, &policy.MinTeamSize, &policy.MaxTeamSize)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", models.TeamPolicy{}, fmt.Errorf("hackathon not found: %w", ErrNotFound)
	}
	if err != nil {
		return "", "", models.TeamPolicy{}, mapSQLError(err)
	}
	policy.HackathonID = hackathonID
	if ruleID.Valid {
		return state, ruleID.String, policy, nil
	}
	return state, "", policy, nil
}

func mergeMetadata(existing json.RawMessage, patch json.RawMessage) (json.RawMessage, error) {
	if len(patch) == 0 {
		return normalizeMetadata(existing), nil
	}
	if len(existing) == 0 {
		return normalizeMetadata(patch), nil
	}

	var base map[string]any
	if err := json.Unmarshal(existing, &base); err != nil {
		return nil, err
	}
	var delta map[string]any
	if err := json.Unmarshal(patch, &delta); err != nil {
		return nil, err
	}

	for k, v := range delta {
		base[k] = v
	}
	out, err := json.Marshal(base)
	if err != nil {
		return nil, err
	}
	return out, nil
}
