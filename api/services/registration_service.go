package services

import (
	"context"
	"database/sql"
	"errors"
	
	"github.com/DataInCube/hackathon-service/internal/models"
)

type RegistrationService struct {
	DB *sql.DB
}

func NewRegistrationService(db *sql.DB) *RegistrationService {
	return &RegistrationService{DB: db}
}

func (s *RegistrationService) Register(ctx context.Context, r models.Registration) error {
	query := `INSERT INTO registrations (participant_id, hackathon_id, team_id, created_at, updated_at)
	          VALUES ($1, $2, $3, NOW(), NOW())`

	_, err := s.DB.ExecContext(ctx, query, r.ParticipantID, r.HackathonID, r.TeamID)
	return err
}

func (s *RegistrationService) GetAll(ctx context.Context) ([]models.Registration, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id, participant_id, hackathon_id, team_id, created_at, updated_at FROM registrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regs []models.Registration
	for rows.Next() {
		var r models.Registration
		if err := rows.Scan(&r.ID, &r.ParticipantID, &r.HackathonID, &r.TeamID, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		regs = append(regs, r)
	}
	return regs, nil
}

func (s *RegistrationService) GetByID(ctx context.Context, id uint) (*models.Registration, error) {
	query := `SELECT id, participant_id, hackathon_id, team_id, created_at, updated_at FROM registrations WHERE id = $1`
	row := s.DB.QueryRowContext(ctx, query, id)

	var r models.Registration
	if err := row.Scan(&r.ID, &r.ParticipantID, &r.HackathonID, &r.TeamID, &r.CreatedAt, &r.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (s *RegistrationService) Delete(ctx context.Context, id uint) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM registrations WHERE id = $1`, id)
	return err
}