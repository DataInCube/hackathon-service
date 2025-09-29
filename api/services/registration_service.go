package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DataInCube/hackathon-service/internal/models"
)

type RegistrationService struct {
	DB *sql.DB
}

func NewRegistrationService(db *sql.DB) *RegistrationService {
	return &RegistrationService{DB: db}
}

func (s *RegistrationService) Register(ctx context.Context, r models.Registration) (int64, error) {
	query := `INSERT INTO registrations (participant_id, hackathon_id, team_id, created_at, updated_at)
	          VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`

	var id int64
	err := s.DB.QueryRowContext(ctx, query, r.HackathonID, r.ParticipantID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *RegistrationService) GetAllRegistrations(ctx context.Context) ([]models.Registration, error) {
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

func (s *RegistrationService) GetRegistrationByID(ctx context.Context, id uint) (*models.Registration, error) {
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

func (s *RegistrationService) UpdateRegistration(ctx context.Context, id uint, r models.Registration) error {
	_, err := s.DB.ExecContext(ctx, `UPDATE registrations SET participant_id = $1, hackathon_id = $2, team_id = $3, updated_at = NOW() WHERE id = $4`, r.ParticipantID, r.HackathonID, r.TeamID, id)
	return err
}

func (s *RegistrationService) DeleteRegistration(ctx context.Context, id uint) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM registrations WHERE id = $1`, id)
	return err
}

func (s *RegistrationService) RegisterIndividual(ctx context.Context, participantID, hackathonID uint) (uint, error) {
	// Vérifier si déjà inscrit
	var existingID uint
	err := s.DB.QueryRowContext(ctx, `SELECT id FROM registrations WHERE participant_id=$1 AND hackathon_id=$2`, participantID, hackathonID).Scan(&existingID)
	if err == nil {
		return 0, fmt.Errorf("participant already registered")
	}

	query := `INSERT INTO registrations (participant_id, hackathon_id, created_at, updated_at)
	          VALUES ($1, $2, NOW(), NOW()) RETURNING id`

	var regID uint
	err = s.DB.QueryRowContext(ctx, query, participantID, hackathonID).Scan(&regID)
	if err != nil {
		return 0, err
	}
	return regID, nil
}

func (s *RegistrationService) RegisterToTeam(ctx context.Context, participantID, hackathonID, teamID uint) (uint, error) {
	// Vérifier si la team existe
	var count int
	err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM teams WHERE id=$1 AND hackathon_id=$2`, teamID, hackathonID).Scan(&count)
	if err != nil || count == 0 {
		return 0, fmt.Errorf("team not found for this hackathon")
	}

	// Vérifier si déjà inscrit
	var existingID uint
	err = s.DB.QueryRowContext(ctx, `SELECT id FROM registrations WHERE participant_id=$1 AND hackathon_id=$2`, participantID, hackathonID).Scan(&existingID)
	if err == nil {
		return 0, fmt.Errorf("participant already registered")
	}

	query := `INSERT INTO registrations (participant_id, hackathon_id, team_id, created_at, updated_at)
	          VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`

	var regID uint
	err = s.DB.QueryRowContext(ctx, query, participantID, hackathonID, teamID).Scan(&regID)
	if err != nil {
		return 0, err
	}
	return regID, nil
}

func (s *RegistrationService) ApproveTeamJoin(ctx context.Context, teamLeadID, participantID, teamID uint) error {
	// Vérifier si teamLeadID est bien le leader de cette team
	var leaderID uint
	err := s.DB.QueryRowContext(ctx, `SELECT leader_id FROM teams WHERE id=$1`, teamID).Scan(&leaderID)
	if err != nil {
		return fmt.Errorf("team not found")
	}
	if leaderID != teamLeadID {
		return errors.New("only team leader can approve requests")
	}

	// Mettre à jour la registration
	query := `UPDATE registrations SET team_id=$1, updated_at=NOW()
	          WHERE participant_id=$2 AND team_id IS NULL`
	_, err = s.DB.ExecContext(ctx, query, teamID, participantID)
	if err != nil {
		return err
	}
	return nil
}

func (s *RegistrationService) GetRegistrationsByHackathon(ctx context.Context, hackathonID uint) ([]models.Registration, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id, participant_id, hackathon_id, team_id, created_at, updated_at FROM registrations WHERE hackathon_id=$1`, hackathonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regs []models.Registration
	for rows.Next() {
		var r models.Registration
		err := rows.Scan(&r.ID, &r.ParticipantID, &r.HackathonID, &r.TeamID, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			return nil, err
		}
		regs = append(regs, r)
	}
	return regs, nil
}

func (s *RegistrationService) GetRegistrationByParticipant(ctx context.Context, participantID, hackathonID uint) (*models.Registration, error) {
	query := `SELECT id, participant_id, hackathon_id, team_id, created_at, updated_at
	          FROM registrations WHERE participant_id=$1 AND hackathon_id=$2`

	var r models.Registration
	err := s.DB.QueryRowContext(ctx, query, participantID, hackathonID).
		Scan(&r.ID, &r.ParticipantID, &r.HackathonID, &r.TeamID, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}