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

func setupTeamEnv(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *handlers.TeamHandler) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock error: %v", err)
	}
	service := services.NewTeamService(db)
	h := handlers.NewTeamHandler(service)
	return db, mock, h
}

func TestTeamHandler_CRUD(t *testing.T) {
	db, mock, h := setupTeamEnv(t)
	defer db.Close()
	e := echo.New()

	team := models.Team{Name: "Alpha", HackathonID: 1, LeadID: 1}

	// Create
	mock.ExpectExec("INSERT INTO teams").
		WithArgs(team.Name, team.HackathonID, team.LeadID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(team)
	req := httptest.NewRequest(http.MethodPost, "/api/teams", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// List
	rows := sqlmock.NewRows([]string{"id", "name", "hackathon_id", "lead_id", "created_at", "updated_at"}).
		AddRow(1, team.Name, team.HackathonID, team.LeadID, "now", "now")
	mock.ExpectQuery("SELECT id, name, hackathon_id, lead_id, created_at, updated_at FROM teams").
		WillReturnRows(rows)

	req = httptest.NewRequest(http.MethodGet, "/api/teams", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = h.List(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// GetByID
	row := sqlmock.NewRows([]string{"id", "name", "hackathon_id", "lead_id", "created_at", "updated_at"}).
		AddRow(1, team.Name, team.HackathonID, team.LeadID, "now", "now")
	mock.ExpectQuery("SELECT id, name, hackathon_id, lead_id, created_at, updated_at FROM teams WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(row)

	req = httptest.NewRequest(http.MethodGet, "/api/teams/1", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err = h.GetByID(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Update
	mock.ExpectExec("UPDATE teams SET").
		WithArgs("Alpha Updated", team.HackathonID, team.LeadID, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	update := models.Team{Name: "Alpha Updated", HackathonID: team.HackathonID, LeadID: team.LeadID}
	body, _ = json.Marshal(update)
	req = httptest.NewRequest(http.MethodPut, "/api/teams/1", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err = h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Delete
	mock.ExpectExec("DELETE FROM teams WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req = httptest.NewRequest(http.MethodDelete, "/api/teams/1", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err = h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	assert.NoError(t, mock.ExpectationsWereMet())
}
