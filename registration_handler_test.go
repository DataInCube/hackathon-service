package unit_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/labstack/echo/v4"

	"github.com/DataInCube/hackathon-service/api/handlers"
	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/DataInCube/hackathon-service/internal/models"
)

func setupRegistrationEnv(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *handlers.RegistrationHandler) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock error: %v", err)
	}
	service := services.NewRegistrationService(db)
	h := handlers.NewRegistrationHandler(service)
	return db, mock, h
}

func TestRegistrationHandler_CRUD(t *testing.T) {
	db, mock, h := setupRegistrationEnv(t)
	defer db.Close()
	e := echo.New()

	reg := models.Registration{ParticipantID: 1, HackathonID: 1, TeamID: nil}

	// Create
	mock.ExpectExec("INSERT INTO registrations").
		WithArgs(reg.ParticipantID, reg.HackathonID, reg.TeamID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(reg)
	req := httptest.NewRequest(http.MethodPost, "/api/registrations", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Register(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// List
	rows := sqlmock.NewRows([]string{"id", "participant_id", "hackathon_id", "team_id", "created_at", "updated_at"}).
		AddRow(1, reg.ParticipantID, reg.HackathonID, nil, "now", "now")
	mock.ExpectQuery("SELECT id, participant_id, hackathon_id, team_id, created_at, updated_at FROM registrations").
		WillReturnRows(rows)

	req = httptest.NewRequest(http.MethodGet, "/api/registrations", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = h.List(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// GetByID
	row := sqlmock.NewRows([]string{"id", "participant_id", "hackathon_id", "team_id", "created_at", "updated_at"}).
		AddRow(1, reg.ParticipantID, reg.HackathonID, nil, "now", "now")
	mock.ExpectQuery("SELECT id, participant_id, hackathon_id, team_id, created_at, updated_at FROM registrations WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(row)

	req = httptest.NewRequest(http.MethodGet, "/api/registrations/1", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err = h.GetByID(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Delete
	mock.ExpectExec("DELETE FROM registrations WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req = httptest.NewRequest(http.MethodDelete, "/api/registrations/1", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err = h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	assert.NoError(t, mock.ExpectationsWereMet())
}
