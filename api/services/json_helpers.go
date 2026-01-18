package services

import (
	"encoding/json"
	"fmt"
)

func ensureValidJSON(raw json.RawMessage, field string) error {
	if len(raw) == 0 {
		return nil
	}
	if !json.Valid(raw) {
		if field == "" {
			field = "json"
		}
		return fmt.Errorf("invalid %s: %w", field, ErrInvalid)
	}
	return nil
}
