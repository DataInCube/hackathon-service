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

func setupParticipantEnv(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *handlers.ParticipantHandler) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock error: %v", err)
	}
	service := services.NewParticipantService(db)
	h := handlers.NewParticipantHandler(service)
	return db, mock, h
}

func TestParticipantHandler_CRUD(t *testing.T) {
	db, mock, h := setupParticipantEnv(t)
	defer db.Close()
	e := echo.New()

	p := models.Participant{Name: "John Doe", Email: "john@example.com", UserID: "kc-1"}

	// Create
	mock.ExpectExec("INSERT INTO participants").
		WithArgs(p.Name, p.Email, p.UserID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(p)
	req := httptest.NewRequest(http.MethodPost, "/api/participants", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// List
	rows := sqlmock.NewRows([]string{"id", "name", "email", "user_id", "created_at", "updated_at"}).
		AddRow(1, p.Name, p.Email, p.UserID, "now", "now")
	mock.ExpectQuery("SELECT id, name, email, user_id, created_at, updated_at FROM participants").
		WillReturnRows(rows)

	req = httptest.NewRequest(http.MethodGet, "/api/participants", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = h.List(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// GetByID
	row := sqlmock.NewRows([]string{"id", "name", "email", "user_id", "created_at", "updated_at"}).
		AddRow(1, p.Name, p.Email, p.UserID, "now", "now")
	mock.ExpectQuery("SELECT id, name, email, user_id, created_at, updated_at FROM participants WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(row)

	req = httptest.NewRequest(http.MethodGet, "/api/participants/1", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err = h.GetByID(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Update
	mock.ExpectExec("UPDATE participants SET").
		WithArgs("John Updated", p.Email, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	update := models.Participant{Name: "John Updated", Email: p.Email}
	body, _ = json.Marshal(update)
	req = httptest.NewRequest(http.MethodPut, "/api/participants/1", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err = h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Delete
	mock.ExpectExec("DELETE FROM participants WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req = httptest.NewRequest(http.MethodDelete, "/api/participants/1", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err = h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	assert.NoError(t, mock.ExpectationsWereMet())
}
