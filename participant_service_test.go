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

func setupParticipantService(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *services.ParticipantService) {
	db, mock, _ := sqlmock.New()
	return db, mock, services.NewParticipantService(db)
}

func TestParticipantService_CRUD(t *testing.T) {
	db, mock, service := setupParticipantService(t)
	defer db.Close()

	p := &models.Participant{Name: "Alice", Email: "alice@example.com"}

	// Create
	mock.ExpectQuery("INSERT INTO participants").
		WithArgs(p.Name, p.Email).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	id, err := service.Create(context.Background(), *p)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)

	// GetByID
	mock.ExpectQuery("SELECT id, name, email FROM participants WHERE id").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).AddRow(1, p.Name, p.Email))

	got, err := service.GetByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "Alice", got.Name)

	// Update
	mock.ExpectExec("UPDATE participants SET").
		WithArgs("Alice Updated", "alice@example.com", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = service.Update(context.Background(), *p)
	assert.NoError(t, err)

	// Delete
	mock.ExpectExec("DELETE FROM participants WHERE id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = service.Delete(context.Background(), 1)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
