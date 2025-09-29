package unit_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/labstack/echo/v4"

	"github.com/DataInCube/hackathon-service/api/handlers"
	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/DataInCube/hackathon-service/internal/models"
)

func setupHackathonEnv(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *handlers.HackathonHandler) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	service := services.NewHackathonService(db)
	h := handlers.NewHackathonHandler(service)
	return db, mock, h
}

func TestHackathonHandler_Create_List_Get_Update_Delete(t *testing.T) {
	db, mock, h := setupHackathonEnv(t)
	defer db.Close()

	e := echo.New()

	// --- Create ---
	hack := models.Hackathon{
		Title:       "AI Hackathon",
		Description: "Test AI",
		StartDate:   time.Now(),
		EndDate:     time.Now().Add(48 * time.Hour),
	}

	// Expect Exec (INSERT)
	mock.ExpectExec("INSERT INTO hackathons").
		WithArgs(hack.Title, hack.Description, hack.StartDate, hack.EndDate).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(hack)
	req := httptest.NewRequest(http.MethodPost, "/api/hackathons", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// --- List ---
	rows := sqlmock.NewRows([]string{"id", "title", "description", "start_date", "end_date", "created_at", "updated_at"}).
		AddRow(1, hack.Title, hack.Description, hack.StartDate, hack.EndDate, time.Now(), time.Now())

	mock.ExpectQuery("SELECT id, title, description, start_date, end_date, created_at, updated_at FROM hackathons").
		WillReturnRows(rows)

	req = httptest.NewRequest(http.MethodGet, "/api/hackathons", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = h.List(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// --- GetByID ---
	row := sqlmock.NewRows([]string{"id", "title", "description", "start_date", "end_date", "created_at", "updated_at"}).
		AddRow(1, hack.Title, hack.Description, hack.StartDate, hack.EndDate, time.Now(), time.Now())

	mock.ExpectQuery("SELECT id, title, description, start_date, end_date, created_at, updated_at FROM hackathons WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(row)

	req = httptest.NewRequest(http.MethodGet, "/api/hackathons/1", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err = h.GetByID(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// --- Update ---
	mock.ExpectExec("UPDATE hackathons SET").
		WithArgs("Updated Title", hack.Description, hack.StartDate, hack.EndDate, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	updatePayload := models.Hackathon{Title: "Updated Title", Description: hack.Description, StartDate: hack.StartDate, EndDate: hack.EndDate}
	body, _ = json.Marshal(updatePayload)
	req = httptest.NewRequest(http.MethodPut, "/api/hackathons/1", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err = h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// --- Delete ---
	mock.ExpectExec("DELETE FROM hackathons WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req = httptest.NewRequest(http.MethodDelete, "/api/hackathons/1", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err = h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// ensure expectations met
	assert.NoError(t, mock.ExpectationsWereMet())
}
