package unit_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/stretchr/testify/assert"
)

func setupTeamService(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *services.TeamService) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	service := services.NewTeamService(db, nil)
	return db, mock, service
}

func TestTeamService_CRUD(t *testing.T) {
	db, mock, service := setupTeamService(t)
	defer db.Close()

	team := &models.Team{Name: "Dream Team", HackathonID: 1, LeadID: 1, RepoURL: "https://github.com/dream-team"}

	// --- Create ---
	mock.ExpectQuery("INSERT INTO teams").
		WithArgs(team.Name, team.HackathonID, team.LeadID, team.RepoURL).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	id, err := service.CreateTeam(context.Background(), *team, []string{})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)

	// --- GetByID ---
	mock.ExpectQuery("SELECT id, name, hackathon_id, lead_id, repo_url, created_at, updated_at FROM teams WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "hackathon_id", "lead_id", "repo_url", "created_at", "updated_at"}).
			AddRow(1, team.Name, team.HackathonID, team.LeadID, team.RepoURL, "2025-10-01", "2025-10-01"))

	got, err := service.GetTeamByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "Dream Team", got.Name)

	// --- Update ---
	updatedTeam := &models.Team{ID: 1, Name: "Updated Team", HackathonID: team.HackathonID, LeadID: team.LeadID}
	mock.ExpectExec("UPDATE teams SET").
		WithArgs(updatedTeam.Name, updatedTeam.HackathonID, updatedTeam.LeadID, updatedTeam.ID).
		WillReturnResult(sqlmock.NewResult(0, 1)) // rows affected = 1

	err = service.UpdateTeam(context.Background(), *updatedTeam)
	assert.NoError(t, err)

	// --- Delete ---
	mock.ExpectExec("DELETE FROM teams WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1)) // rows affected = 1

	err = service.DeleteTeam(context.Background(), 1)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
