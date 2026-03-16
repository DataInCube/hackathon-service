package utils

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestNewDB_InvalidDriver(t *testing.T) {
	logger := logrus.New()
	db, err := NewDB("invalid-driver", "dsn", 1, 1, time.Second, logger)
	if err == nil {
		t.Fatal("expected error for invalid driver")
	}
	if db != nil {
		t.Fatal("db should be nil when open fails")
	}
}
