package consumer

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/microservices/shared/messaging"
)

const (
	QueueUserEvents  = "notifications.user.events"
	QueueOrderEvents = "notifications.order.events"
)

type Consumer struct {
	conn *messaging.Connection
	log  *slog.Logger
}

func New(conn *messaging.Connection, log *slog.Logger) *Consumer {
	return &Consumer{conn: conn, log: log}
}

func (c *Consumer) Setup() error {
	// Bind user events
	if _, err := c.conn.DeclareQueue(QueueUserEvents); err != nil {
		return fmt.Errorf("declare user queue: %w", err)
	}
	for _, key := range []string{"user.created", "user.updated", "user.deleted"} {
		if err := c.conn.BindQueue(QueueUserEvents, "users.events", key); err != nil {
			return fmt.Errorf("bind user queue (%s): %w", key, err)
		}
	}

	// Bind order events
	if _, err := c.conn.DeclareQueue(QueueOrderEvents); err != nil {
		return fmt.Errorf("declare order queue: %w", err)
	}
	for _, key := range []string{"order.created", "order.status_changed", "order.cancelled"} {
		if err := c.conn.BindQueue(QueueOrderEvents, "orders.events", key); err != nil {
			return fmt.Errorf("bind order queue (%s): %w", key, err)
		}
	}

	return nil
}

func (c *Consumer) Start() error {
	if err := c.conn.Consume(QueueUserEvents, c.handleUserEvent); err != nil {
		return fmt.Errorf("consume user events: %w", err)
	}

	if err := c.conn.Consume(QueueOrderEvents, c.handleOrderEvent); err != nil {
		return fmt.Errorf("consume order events: %w", err)
	}

	c.log.Info("notification consumer started", "queues", []string{QueueUserEvents, QueueOrderEvents})
	return nil
}

func (c *Consumer) handleUserEvent(msg messaging.Message) error {
	payload, _ := json.Marshal(msg.Payload)

	switch msg.Type {
	case "user.created":
		c.log.Info("sending welcome email", "event", msg.Type, "payload", string(payload))
		// TODO: integrate with email provider (SendGrid, SES, etc.)

	case "user.updated":
		c.log.Info("sending account update notification", "event", msg.Type)

	case "user.deleted":
		c.log.Info("sending account deactivation email", "event", msg.Type)

	default:
		c.log.Warn("unknown user event", "type", msg.Type)
	}

	return nil
}

func (c *Consumer) handleOrderEvent(msg messaging.Message) error {
	payload, _ := json.Marshal(msg.Payload)

	switch msg.Type {
	case "order.created":
		c.log.Info("sending order confirmation", "event", msg.Type, "payload", string(payload))
		// TODO: send order confirmation email/SMS

	case "order.status_changed":
		c.log.Info("sending order status update", "event", msg.Type, "payload", string(payload))
		// TODO: send push notification / email

	case "order.cancelled":
		c.log.Info("sending order cancellation notice", "event", msg.Type, "payload", string(payload))
		// TODO: send cancellation email + trigger refund flow

	default:
		c.log.Warn("unknown order event", "type", msg.Type)
	}

	return nil
}
