package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/DataInCube/hackathon-service/internal/models"
)

type MessageService struct {
	DB *sql.DB
}

func NewMessageService(db *sql.DB) *MessageService {
	return &MessageService{DB: db}
}

func (s *MessageService) Create(ctx context.Context, m models.Message) (int64, error) {
	query := `INSERT INTO messages (sender_id, team_id, hackathon_id, content, created_at)
	          VALUES ($1, $2, $3, $4, $5) RETURNING id`

	var id int64
	err := s.DB.QueryRowContext(ctx, query, m.SenderID, m.TeamID, m.HackathonID, m.Content, time.Now()).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *MessageService) GetByID(ctx context.Context, id uint) (*models.Message, error) {
	query := `SELECT id, sender_id, team_id, hackathon_id, content, created_at FROM messages WHERE id = $1`
	row := s.DB.QueryRowContext(ctx, query, id)

	var m models.Message
	err := row.Scan(&m.ID, &m.SenderID, &m.TeamID, &m.HackathonID, &m.Content, &m.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &m, err
}

func (s *MessageService) GetByHackathon(ctx context.Context, hackathonID uint) ([]models.Message, error) {
	query := `SELECT id, sender_id, team_id, hackathon_id, content, created_at 
	          FROM messages WHERE hackathon_id = $1 ORDER BY created_at ASC`

	rows, err := s.DB.QueryContext(ctx, query, hackathonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.TeamID, &m.HackathonID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (s *MessageService) GetByTeam(ctx context.Context, teamID uint) ([]models.Message, error) {
	query := `SELECT id, sender_id, team_id, hackathon_id, content, created_at 
	          FROM messages WHERE team_id = $1 ORDER BY created_at ASC`

	rows, err := s.DB.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.TeamID, &m.HackathonID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (s *MessageService) List(ctx context.Context) ([]models.Message, error) {
	query := `SELECT id, sender_id, team_id, hackathon_id, content, created_at FROM messages ORDER BY created_at ASC`

	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.TeamID, &m.HackathonID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (s *MessageService) Delete(ctx context.Context, id uint) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM messages WHERE id = $1`, id)
	return err
}