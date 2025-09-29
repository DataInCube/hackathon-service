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

func setupRegistrationService(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *services.RegistrationService) {
	db, mock, _ := sqlmock.New()
	return db, mock, services.NewRegistrationService(db)
}

func TestRegistrationService_CRUD(t *testing.T) {
	db, mock, service := setupRegistrationService(t)
	defer db.Close()

	r := &models.Registration{HackathonID: uint(1), ParticipantID: uint(2), TeamID: nil}

	// Create
	mock.ExpectQuery("INSERT INTO registrations").
		WithArgs(r.HackathonID, r.ParticipantID, r.TeamID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint(1)))

	id, err := service.Register(context.Background(), *r)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)

	// GetByID
	mock.ExpectQuery("SELECT id, hackathon_id, participant_id, team_id FROM registrations WHERE id").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "hackathon_id", "participant_id", "team_id"}).AddRow(1, 1, 2, nil))

	got, err := service.GetRegistrationByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), got.ParticipantID)

	// Update
	mock.ExpectExec("UPDATE registrations SET").
		WithArgs(1, 2, 2, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = service.UpdateRegistration(context.Background(), 1, *r)
	assert.NoError(t, err)

	// Delete
	mock.ExpectExec("DELETE FROM registrations WHERE id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = service.DeleteRegistration(context.Background(), 1)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
