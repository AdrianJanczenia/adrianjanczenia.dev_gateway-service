package rabbitmq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn     *amqp091.Connection
	exchange string
}

func NewClient(url, exchange string) (*Client, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}

	return &Client{conn: conn, exchange: exchange}, nil
}

func (c *Client) Request(routingKey string, payload any) (body []byte, err error) {
	ch, err := c.conn.Channel()
	if err != nil {
		return nil, errors.ErrInternalServerError
	}
	defer ch.Close()

	replyQueue, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		return nil, errors.ErrInternalServerError
	}

	msgs, err := ch.Consume(replyQueue.Name, "", true, true, false, false, nil)
	if err != nil {
		return nil, errors.ErrInternalServerError
	}

	corrID := uuid.New().String()
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.ErrInternalServerError
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx, c.exchange, routingKey, false, false, amqp091.Publishing{
		ContentType:   "application/json",
		CorrelationId: corrID,
		ReplyTo:       replyQueue.Name,
		Body:          requestBody,
	})
	if err != nil {
		return nil, errors.ErrServiceUnavailable
	}

	select {
	case d := <-msgs:
		if corrID == d.CorrelationId {
			var resp struct {
				Error string `json:"error"`
			}
			json.Unmarshal(d.Body, &resp)
			if resp.Error != "" {
				return nil, errors.FromSlug(resp.Error)
			}
			return d.Body, nil
		}
	case <-time.After(5 * time.Second):
		return nil, errors.ErrServiceUnavailable
	}

	return nil, errors.ErrInternalServerError
}

func (c *Client) Close() error {
	return c.conn.Close()
}
