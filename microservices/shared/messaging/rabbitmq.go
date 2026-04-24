package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Connection struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	log     *slog.Logger
}

type Message struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Payload   any       `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

func Connect(url string, log *slog.Logger) (*Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	// Enable publisher confirms
	if err := ch.Confirm(false); err != nil {
		return nil, fmt.Errorf("confirm mode: %w", err)
	}

	return &Connection{conn: conn, channel: ch, log: log}, nil
}

func (c *Connection) DeclareExchange(name, kind string) error {
	return c.channel.ExchangeDeclare(name, kind, true, false, false, false, nil)
}

func (c *Connection) DeclareQueue(name string) (amqp.Queue, error) {
	return c.channel.QueueDeclare(name, true, false, false, false, nil)
}

func (c *Connection) BindQueue(queue, exchange, routingKey string) error {
	return c.channel.QueueBind(queue, routingKey, exchange, false, nil)
}

func (c *Connection) Publish(ctx context.Context, exchange, routingKey string, msg Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	err = c.channel.PublishWithContext(ctx, exchange, routingKey, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish message: %w", err)
	}

	c.log.Info("message published", "exchange", exchange, "routing_key", routingKey, "type", msg.Type)
	return nil
}

func (c *Connection) Consume(queue string, handler func(Message) error) error {
	msgs, err := c.channel.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume queue %s: %w", queue, err)
	}

	go func() {
		for d := range msgs {
			var msg Message
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				c.log.Error("unmarshal message", "error", err)
				d.Nack(false, false)
				continue
			}

			if err := handler(msg); err != nil {
				c.log.Error("handle message", "type", msg.Type, "error", err)
				d.Nack(false, true) // requeue
				continue
			}

			d.Ack(false)
			c.log.Info("message processed", "type", msg.Type, "id", msg.ID)
		}
	}()

	return nil
}

func (c *Connection) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
