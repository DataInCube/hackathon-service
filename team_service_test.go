package unit_test
import (
	"testing"
	"context"
	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/DataInCube/hackathon-service/api/services"
)

func setupTeamService(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *services.TeamService) {	
	db, mock, _ := sqlmock.New()
	return db, mock, services.NewTeamService(db)
}

func TestTeamService_CRUD(t *testing.T) {
	db, mock, service := setupTeamService(t)
	defer db.Close()

	team := &models.Team{Name: "Dream Team", HackathonID: 1}

	// Create
	mock.ExpectQuery("INSERT INTO teams").
		WithArgs(team.Name, team.HackathonID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	id, err := service.CreateTeam(context.Background(), *team)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)

	// GetByID
	mock.ExpectQuery("SELECT id, name, hackathon_id FROM teams WHERE id").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "hackathon_id"}).AddRow(1, team.Name, team.HackathonID))

	got, err := service.GetTeamByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "Dream Team", got.Name)

	// Update
	mock.ExpectExec("UPDATE teams SET").
		WithArgs("Updated Team", 1, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = service.UpdateTeam(context.Background(), *team)
	assert.NoError(t, err)

	// Delete
	mock.ExpectExec("DELETE FROM teams WHERE id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = service.DeleteTeam(context.Background(), 1)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}