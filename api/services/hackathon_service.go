package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DataInCube/hackathon-service/internal/models"
)

type HackathonService struct {
	DB *sql.DB
}

func NewHackathonService(db *sql.DB) *HackathonService {
	return &HackathonService{DB: db}
}

func (s *HackathonService) CreateHackathon(ctx context.Context, h models.Hackathon) (int64, error) {
	query := `INSERT INTO hackathons (title, description, start_date, end_date, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING id`

	var id int64
	err := s.DB.QueryRowContext(ctx, query, h.Title, h.Description, h.StartDate, h.EndDate).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *HackathonService) GetAllHackathons(ctx context.Context) ([]models.Hackathon, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id, title, description, start_date, end_date, created_at, updated_at FROM hackathons`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hackathons []models.Hackathon
	for rows.Next() {
		var h models.Hackathon
		if err := rows.Scan(&h.ID, &h.Title, &h.Description, &h.StartDate, &h.EndDate, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		hackathons = append(hackathons, h)
	}
	return hackathons, nil
}

func (s *HackathonService) GetHackathonByID(ctx context.Context, id uint) (*models.Hackathon, error) {
	query := `SELECT id, title, description, start_date, end_date, created_at, updated_at FROM hackathons WHERE id = $1`
	row := s.DB.QueryRowContext(ctx, query, id)

	var h models.Hackathon
	if err := row.Scan(&h.ID, &h.Title, &h.Description, &h.StartDate, &h.EndDate, &h.CreatedAt, &h.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &h, nil
}

func (s *HackathonService) UpdateHackathon(ctx context.Context, h models.Hackathon) error {
	query := `UPDATE hackathons SET title = $1, description = $2, start_date = $3, end_date = $4, updated_at = NOW() WHERE id = $5`
	_, err := s.DB.ExecContext(ctx, query, h.Title, h.Description, h.StartDate, h.EndDate, h.ID)
	return err
}

func (s *HackathonService) DeleteHackathon(ctx context.Context, id uint) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM hackathons WHERE id = $1`, id)
	return err
}
