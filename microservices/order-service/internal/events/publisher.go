package events

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/microservices/shared/messaging"
	"github.com/microservices/order-service/internal/domain"
)

const (
	ExchangeOrders = "orders.events"

	EventOrderCreated        = "order.created"
	EventOrderStatusChanged  = "order.status_changed"
	EventOrderCancelled      = "order.cancelled"
)

type OrderPayload struct {
	ID          string              `json:"id"`
	UserID      string              `json:"user_id"`
	TotalAmount float64             `json:"total_amount"`
	Status      domain.OrderStatus  `json:"status"`
	ItemCount   int                 `json:"item_count"`
}

type StatusChangedPayload struct {
	OrderID   string             `json:"order_id"`
	UserID    string             `json:"user_id"`
	OldStatus domain.OrderStatus `json:"old_status"`
	NewStatus domain.OrderStatus `json:"new_status"`
}

type Publisher struct {
	conn *messaging.Connection
}

func NewPublisher(conn *messaging.Connection) *Publisher {
	return &Publisher{conn: conn}
}

func (p *Publisher) Setup() error {
	return p.conn.DeclareExchange(ExchangeOrders, "topic")
}

func (p *Publisher) OrderCreated(ctx context.Context, order *domain.Order) error {
	return p.conn.Publish(ctx, ExchangeOrders, EventOrderCreated, messaging.Message{
		ID:        uuid.NewString(),
		Type:      EventOrderCreated,
		Payload:   OrderPayload{
			ID:          order.ID,
			UserID:      order.UserID,
			TotalAmount: order.TotalAmount,
			Status:      order.Status,
			ItemCount:   len(order.Items),
		},
		Timestamp: time.Now(),
	})
}

func (p *Publisher) StatusChanged(ctx context.Context, payload StatusChangedPayload) error {
	eventType := EventOrderStatusChanged
	if payload.NewStatus == domain.StatusCancelled {
		eventType = EventOrderCancelled
	}

	return p.conn.Publish(ctx, ExchangeOrders, eventType, messaging.Message{
		ID:        uuid.NewString(),
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now(),
	})
}
