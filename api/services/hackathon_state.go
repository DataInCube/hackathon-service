package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DataInCube/hackathon-service/internal/models"
)

func loadHackathonState(ctx context.Context, db *sql.DB, hackathonID string) (string, error) {
	var state string
	err := db.QueryRowContext(ctx, `SELECT state FROM hackathons WHERE id = $1`, hackathonID).Scan(&state)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("hackathon not found: %w", ErrNotFound)
	}
	if err != nil {
		return "", mapSQLError(err)
	}
	return state, nil
}

func ensureEditableHackathon(ctx context.Context, db *sql.DB, hackathonID string) error {
	state, err := loadHackathonState(ctx, db, hackathonID)
	if err != nil {
		return err
	}
	if !isHackathonEditable(state) {
		return fmt.Errorf("hackathon not editable in state %s: %w", state, ErrInvalid)
	}
	return nil
}

func isHackathonEditable(state string) bool {
	return state == models.HackathonStateDraft || state == models.HackathonStatePublished
}
