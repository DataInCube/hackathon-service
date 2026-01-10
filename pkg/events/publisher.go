package events

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type Publisher interface {
	Publish(ctx context.Context, subject string, payload any) error
	Close()
}

type NatsPublisher struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

type Envelope struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Source     string      `json:"source"`
	OccurredAt time.Time   `json:"occurred_at"`
	Payload    any         `json:"payload"`
}

func NewNatsPublisher(natsURL string, stream string, subjects []string) (*NatsPublisher, error) {
	if natsURL == "" {
		return nil, errors.New("nats url is required")
	}
	nc, err := nats.Connect(natsURL, nats.Timeout(3*time.Second))
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, err
	}

	if stream != "" && len(subjects) > 0 {
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     stream,
			Subjects: subjects,
			Storage:  nats.FileStorage,
		})
		if err != nil && err != nats.ErrStreamNameAlreadyInUse {
			_, uerr := js.UpdateStream(&nats.StreamConfig{
				Name:     stream,
				Subjects: subjects,
				Storage:  nats.FileStorage,
			})
			if uerr != nil {
				nc.Close()
				return nil, uerr
			}
		}
	}

	return &NatsPublisher{nc: nc, js: js}, nil
}

func (p *NatsPublisher) Publish(ctx context.Context, subject string, payload any) error {
	if p == nil || p.js == nil {
		return nil
	}
	env := Envelope{
		ID:         uuid.NewString(),
		Type:       subject,
		Source:     "hackathon-service",
		OccurredAt: time.Now().UTC(),
		Payload:    payload,
	}
	b, err := json.Marshal(env)
	if err != nil {
		return err
	}
	_, err = p.js.Publish(subject, b)
	return err
}

func (p *NatsPublisher) Close() {
	if p == nil || p.nc == nil {
		return
	}
	p.nc.Close()
}
