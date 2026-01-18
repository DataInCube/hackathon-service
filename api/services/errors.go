package services

import (
	"errors"

	"github.com/lib/pq"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
	ErrInvalid  = errors.New("invalid")
	ErrForbidden = errors.New("forbidden")
)

func mapSQLError(err error) error {
	if err == nil {
		return nil
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch string(pqErr.Code) {
		case "23505":
			return ErrConflict
		case "23503", "23514":
			return ErrInvalid
		default:
			return err
		}
	}

	return err
}
