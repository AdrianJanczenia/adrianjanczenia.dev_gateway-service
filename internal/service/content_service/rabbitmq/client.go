package rabbitmq

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn         *amqp091.Connection
	ch           *amqp091.Channel
	exchange     string
	replyQueue   string
	pendingCalls map[string]chan []byte
	mu           sync.RWMutex
}

func NewClient(url, exchange string) (*Client, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	client := &Client{
		conn:         conn,
		ch:           ch,
		exchange:     exchange,
		replyQueue:   q.Name,
		pendingCalls: make(map[string]chan []byte),
	}

	go client.handleReplies()

	return client, nil
}

func (c *Client) handleReplies() {
	msgs, err := c.ch.Consume(c.replyQueue, "", true, true, false, false, nil)
	if err != nil {
		return
	}

	for d := range msgs {
		c.mu.RLock()
		resChan, ok := c.pendingCalls[d.CorrelationId]
		c.mu.RUnlock()

		if ok {
			resChan <- d.Body
			c.mu.Lock()
			delete(c.pendingCalls, d.CorrelationId)
			c.mu.Unlock()
		}
	}
}

func (c *Client) Request(ctx context.Context, routingKey string, payload any) ([]byte, error) {
	corrID := uuid.New().String()
	resChan := make(chan []byte, 1)

	c.mu.Lock()
	c.pendingCalls[corrID] = resChan
	c.mu.Unlock()

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.ErrInternalServerError
	}

	err = c.ch.PublishWithContext(ctx, c.exchange, routingKey, false, false, amqp091.Publishing{
		ContentType:   "application/json",
		CorrelationId: corrID,
		ReplyTo:       c.replyQueue,
		Body:          requestBody,
	})
	if err != nil {
		c.mu.Lock()
		delete(c.pendingCalls, corrID)
		c.mu.Unlock()
		return nil, errors.ErrServiceUnavailable
	}

	select {
	case body := <-resChan:
		var resp struct {
			Error string `json:"error"`
		}
		json.Unmarshal(body, &resp)
		if resp.Error != "" {
			return nil, errors.FromSlug(resp.Error)
		}
		return body, nil
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pendingCalls, corrID)
		c.mu.Unlock()
		return nil, ctx.Err()
	case <-time.After(10 * time.Second):
		c.mu.Lock()
		delete(c.pendingCalls, corrID)
		c.mu.Unlock()
		return nil, errors.ErrServiceUnavailable
	}
}

func (c *Client) Close() error {
	_ = c.ch.Close()
	return c.conn.Close()
}
