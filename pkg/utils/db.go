package utils

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

func ConnectDB() *sql.DB {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Error opening DB: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Error pinging DB: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL database")
	return db
}

func NewDB(driver, dsn string, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration, logger *logrus.Logger) (*sql.DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening DB: %w", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging DB: %w", err)
	}

	logger.Println("✅ Connected to PostgreSQL database")
	return db, nil
}
