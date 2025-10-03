package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/DataInCube/hackathon-service/pkg/utils"
)

type TeamService struct {
	DB        *sql.DB
	KronosAPI *utils.KronosClient
}

func NewTeamService(db *sql.DB, kronos *utils.KronosClient) *TeamService {
	return &TeamService{
		DB:        db,
		KronosAPI: kronos,
	}
}

// CreateTeam creates a team, a GitHub repo, and adds collaborators
func (s *TeamService) CreateTeam(ctx context.Context, t models.Team, memberGitUsernames []string) (*models.Team, error) {
	// Création DB
	query := `INSERT INTO teams (name, hackathon_id, lead_id, created_at, updated_at)
              VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`
	var id int64
	if err := s.DB.QueryRowContext(ctx, query, t.Name, t.HackathonID, t.LeadID).Scan(&id); err != nil {
		return nil, err
	}
	t.ID = uint(id)

	// Créer le repo via Kronos
	repoName := fmt.Sprintf("team-%d-%s", t.ID, t.Name)
	if err := s.KronosAPI.CreateTeamRepo(repoName); err != nil {
		return nil, fmt.Errorf("erreur création repo: %v", err)
	}

	// Ajouter les collaborateurs
	if len(memberGitUsernames) > 0 {
		if err := s.KronosAPI.AddCollaborators(repoName, memberGitUsernames); err != nil {
			return nil, fmt.Errorf("erreur ajout collaborateurs: %v", err)
		}
	}

	// Ici tu peux stocker l’URL du repo si l’API retourne un champ
	t.RepoURL = fmt.Sprintf("%s/%s", s.KronosAPI.BaseURL, repoName)

	return &t, nil
}

func (s *TeamService) GetAllTeams(ctx context.Context) ([]models.Team, error) {
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

func (s *TeamService) GetTeamByID(ctx context.Context, id uint) (*models.Team, error) {
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

func (s *TeamService) UpdateTeam(ctx context.Context, t models.Team) error {
	query := `UPDATE teams SET name = $1, hackathon_id = $2, lead_id = $3, updated_at = NOW() WHERE id = $4`
	_, err := s.DB.ExecContext(ctx, query, t.Name, t.HackathonID, t.LeadID, t.ID)
	return err
}

func (s *TeamService) DeleteTeam(ctx context.Context, id uint) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM teams WHERE id = $1`, id)
	return err
}

// TransferLead changes the lead of a team to another participant
func (s *TeamService) TransferLead(ctx context.Context, teamID, newLeadID uint) error {
    query := `UPDATE teams SET lead_id = $1, updated_at = NOW() WHERE id = $2`
    _, err := s.DB.ExecContext(ctx, query, newLeadID, teamID)
    return err
}