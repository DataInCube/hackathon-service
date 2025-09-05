package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DataInCube/hackathon-service/internal/models"
)

type TeamService struct {
	DB *sql.DB
}

func NewTeamService(db *sql.DB) *TeamService {
	return &TeamService{DB: db}
}

func (s *TeamService) Create(ctx context.Context, t models.Team) error {
	query := `INSERT INTO teams (name, hackathon_id, lead_id, created_at, updated_at)
	          VALUES ($1, $2, $3, NOW(), NOW())`

	_, err := s.DB.ExecContext(ctx, query, t.Name, t.HackathonID, t.LeadID)
	return err
}

func (s *TeamService) GetAll(ctx context.Context) ([]models.Team, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id, name, hackathon_id, lead_id, created_at, updated_at FROM teams`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var t models.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.HackathonID, &t.LeadID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}

func (s *TeamService) GetByID(ctx context.Context, id uint) (*models.Team, error) {
	row := s.DB.QueryRowContext(ctx, `SELECT id, name, hackathon_id, lead_id, created_at, updated_at FROM teams WHERE id = $1`, id)
	var t models.Team
	if err := row.Scan(&t.ID, &t.Name, &t.HackathonID, &t.LeadID, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (s *TeamService) Update(ctx context.Context, t models.Team) error {
	query := `UPDATE teams SET name = $1, hackathon_id = $2, lead_id = $3, updated_at = NOW() WHERE id = $4`
	_, err := s.DB.ExecContext(ctx, query, t.Name, t.HackathonID, t.LeadID, t.ID)
	return err
}

func (s *TeamService) Delete(ctx context.Context, id uint) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM teams WHERE id = $1`, id)
	return err
}
