package rabbitmq

import (
	"context"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"log/slog"
)

type Client struct {
	conn  *amqp091.Connection
	ch    *amqp091.Channel
	queue string
	log   *slog.Logger
}

func New(url, queue string, log *slog.Logger) (*Client, error) {
	if log == nil {
		log = slog.Default()
	}

	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("channel: %w", err)
	}

	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("queue declare: %w", err)
	}

	log.Info("rabbitmq connected", slog.String("queue", queue))

	return &Client{conn: conn, ch: ch, queue: queue, log: log}, nil
}

func (c *Client) Publish(ctx context.Context, msg []byte) error {
	err := c.ch.PublishWithContext(ctx, "", c.queue, false, false, amqp091.Publishing{
		ContentType: "application/json",
		Body:        msg,
	})
	if err != nil {
		c.log.Error("publish failed", slog.Any("error", err))
		return err
	}
	c.log.Debug("published", slog.Int("size", len(msg)))
	return nil
}

func (c *Client) Channel() (*amqp091.Channel, error) {
	return c.conn.Channel()
}

func (c *Client) Close() error {
	if err := c.ch.Close(); err != nil {
		return err
	}
	return c.conn.Close()
}
