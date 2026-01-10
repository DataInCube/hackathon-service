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

type GovernanceService struct {
	DB *sql.DB
}

func NewGovernanceService(db *sql.DB) *GovernanceService {
	return &GovernanceService{DB: db}
}

func (s *GovernanceService) CreateReport(ctx context.Context, hackathonID string, input models.Report, reporterID string) (*models.Report, error) {
	if input.Type == "" || input.Content == "" {
		return nil, fmt.Errorf("type and content are required: %w", ErrInvalid)
	}

	now := time.Now().UTC()
	report := models.Report{
		ID:          uuid.NewString(),
		HackathonID: hackathonID,
		ReporterID:  reporterID,
		Type:        input.Type,
		Content:     input.Content,
		Status:      "open",
		CreatedAt:   now,
	}

	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO reports (id, hackathon_id, reporter_id, type, content, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		report.ID, report.HackathonID, report.ReporterID, report.Type, report.Content, report.Status, report.CreatedAt,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return &report, nil
}

func (s *GovernanceService) CreateAppeal(ctx context.Context, input models.Appeal, appellantID string) (*models.Appeal, error) {
	if input.SubmissionID == "" || input.Content == "" {
		return nil, fmt.Errorf("submission_id and content are required: %w", ErrInvalid)
	}

	var exists int
	if err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM submissions WHERE id = $1`, input.SubmissionID).Scan(&exists); err != nil {
		return nil, mapSQLError(err)
	}
	if exists == 0 {
		return nil, fmt.Errorf("submission not found: %w", ErrNotFound)
	}

	now := time.Now().UTC()
	appeal := models.Appeal{
		ID:           uuid.NewString(),
		SubmissionID: input.SubmissionID,
		AppellantID:  appellantID,
		Content:      input.Content,
		Status:       "open",
		CreatedAt:    now,
	}

	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO appeals (id, submission_id, appellant_id, content, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		appeal.ID, appeal.SubmissionID, appeal.AppellantID, appeal.Content, appeal.Status, appeal.CreatedAt,
	)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return &appeal, nil
}

func (s *GovernanceService) AuditLogs(ctx context.Context, hackathonID string, limit, offset int) ([]models.AuditLog, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, hackathon_id, actor_id, action, payload, created_at
		FROM audit_logs
		WHERE hackathon_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, hackathonID, limit, offset)
	if err != nil {
		return nil, mapSQLError(err)
	}
	defer rows.Close()

	var items []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		var payload []byte
		if err := rows.Scan(&log.ID, &log.HackathonID, &log.ActorID, &log.Action, &payload, &log.CreatedAt); err != nil {
			return nil, mapSQLError(err)
		}
		log.Payload = payload
		items = append(items, log)
	}
	return items, nil
}

func (s *GovernanceService) AppendAudit(ctx context.Context, log models.AuditLog) error {
	if log.Action == "" {
		return fmt.Errorf("action is required: %w", ErrInvalid)
	}
	if log.ID == "" {
		log.ID = uuid.NewString()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}

	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO audit_logs (id, hackathon_id, actor_id, action, payload, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		log.ID, log.HackathonID, log.ActorID, log.Action, log.Payload, log.CreatedAt,
	)
	if err != nil {
		return mapSQLError(err)
	}
	return nil
}

func (s *GovernanceService) HackathonExists(ctx context.Context, hackathonID string) (bool, error) {
	var count int
	err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM hackathons WHERE id = $1`, hackathonID).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, mapSQLError(err)
	}
	return count > 0, nil
}
