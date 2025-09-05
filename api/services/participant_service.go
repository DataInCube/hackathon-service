package services

import (
	"context"
	"database/sql"
	"errors"
	
	"github.com/DataInCube/hackathon-service/internal/models"
)

type ParticipantService struct {
	DB *sql.DB
}

func NewParticipantService(db *sql.DB) *ParticipantService {
	return &ParticipantService{DB: db}
}

func (s *ParticipantService) Create(ctx context.Context, p models.Participant) error {
	query := `INSERT INTO participants (name, email, user_id, created_at, updated_at)
	          VALUES ($1, $2, $3, NOW(), NOW())`

	_, err := s.DB.ExecContext(ctx, query, p.Name, p.Email, p.UserID)
	return err
}

func (s *ParticipantService) GetByID(ctx context.Context, id uint) (*models.Participant, error) {
	query := `SELECT id, name, email, user_id, created_at, updated_at FROM participants WHERE id = $1`
	row := s.DB.QueryRowContext(ctx, query, id)

	var p models.Participant
	err := row.Scan(&p.ID, &p.Name, &p.Email, &p.UserID, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &p, err
}

func (s *ParticipantService) GetAll(ctx context.Context) ([]models.Participant, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id, name, email, user_id, created_at, updated_at FROM participants`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []models.Participant
	for rows.Next() {
		var p models.Participant
		if err := rows.Scan(&p.ID, &p.Name, &p.Email, &p.UserID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}
	return participants, nil
}

func (s *ParticipantService) Update(ctx context.Context, p models.Participant) error {
	query := `UPDATE participants SET name = $1, email = $2, updated_at = NOW() WHERE id = $3`
	_, err := s.DB.ExecContext(ctx, query, p.Name, p.Email, p.ID)
	return err
}

func (s *ParticipantService) Delete(ctx context.Context, id uint) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM participants WHERE id = $1`, id)
	return err
}