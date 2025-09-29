package unit_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/stretchr/testify/assert"
)

func setupHackathonService(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *services.HackathonService) {
	db, mock, _ := sqlmock.New()
	return db, mock, services.NewHackathonService(db)
}

func TestHackathonService_CRUD(t *testing.T) {
	db, mock, service := setupHackathonService(t)
	defer db.Close()

	h := &models.Hackathon{
		Title:       "AI Hackathon",
		Description: "Test Hackathon",
		StartDate:   time.Now(),
		EndDate:     time.Now().Add(48 * time.Hour),
	}

	// Create
	mock.ExpectQuery("INSERT INTO hackathons").
		WithArgs(h.Title, h.Description, h.StartDate, h.EndDate).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	id, err := service.CreateHackathon(context.Background(), *h)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)

	// GetByID
	mock.ExpectQuery("SELECT id, Title, description, start_date, end_date FROM hackathons WHERE id =").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "Title", "description", "start_date", "end_date"}).
			AddRow(1, h.Title, h.Description, h.StartDate, h.EndDate))

	got, err := service.GetHackathonByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "AI Hackathon", got.Title)

	// Update
	mock.ExpectExec("UPDATE hackathons SET").
		WithArgs("Updated Hackathon", h.Description, h.StartDate, h.EndDate, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = service.UpdateHackathon(context.Background(), models.Hackathon{Title: "Updated Hackathon", Description: h.Description, StartDate: h.StartDate, EndDate: h.EndDate})
	assert.NoError(t, err)

	// Delete
	mock.ExpectExec("DELETE FROM hackathons WHERE id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = service.DeleteHackathon(context.Background(), 1)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
