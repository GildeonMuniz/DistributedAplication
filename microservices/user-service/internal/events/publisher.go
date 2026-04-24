package events

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/microservices/shared/messaging"
)

const (
	ExchangeUsers = "users.events"

	EventUserCreated = "user.created"
	EventUserUpdated = "user.updated"
	EventUserDeleted = "user.deleted"
)

type UserPayload struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type Publisher struct {
	conn *messaging.Connection
}

func NewPublisher(conn *messaging.Connection) *Publisher {
	return &Publisher{conn: conn}
}

func (p *Publisher) Setup() error {
	return p.conn.DeclareExchange(ExchangeUsers, "topic")
}

func (p *Publisher) UserCreated(ctx context.Context, payload UserPayload) error {
	return p.conn.Publish(ctx, ExchangeUsers, EventUserCreated, messaging.Message{
		ID:        uuid.NewString(),
		Type:      EventUserCreated,
		Payload:   payload,
		Timestamp: time.Now(),
	})
}

func (p *Publisher) UserUpdated(ctx context.Context, payload UserPayload) error {
	return p.conn.Publish(ctx, ExchangeUsers, EventUserUpdated, messaging.Message{
		ID:        uuid.NewString(),
		Type:      EventUserUpdated,
		Payload:   payload,
		Timestamp: time.Now(),
	})
}

func (p *Publisher) UserDeleted(ctx context.Context, userID string) error {
	return p.conn.Publish(ctx, ExchangeUsers, EventUserDeleted, messaging.Message{
		ID:        uuid.NewString(),
		Type:      EventUserDeleted,
		Payload:   map[string]string{"id": userID},
		Timestamp: time.Now(),
	})
}
