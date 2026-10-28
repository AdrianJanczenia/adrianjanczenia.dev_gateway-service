package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type reply struct {
	body []byte
	err  error
}

type Client struct {
	url          string
	exchange     string
	mu           sync.Mutex
	conn         *amqp091.Connection
	pub          *amqp091.Channel
	consumer     *amqp091.Channel
	returns      <-chan amqp091.Return
	replyQueue   string
	pendingCalls map[string]chan reply
	inflight     chan struct{}
	closed       bool
}

func NewClient(url, exchange string) (*Client, error) {
	c := &Client{url: url, exchange: exchange, pendingCalls: make(map[string]chan reply), inflight: make(chan struct{}, 64)}
	c.mu.Lock()
	err := c.connectLocked()
	c.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) connectLocked() error {
	if c.closed {
		return fmt.Errorf("RabbitMQ client is closed")
	}
	if c.conn != nil && !c.conn.IsClosed() && c.pub != nil && !c.pub.IsClosed() && c.consumer != nil && !c.consumer.IsClosed() {
		return nil
	}
	c.invalidateLocked()

	conn, err := amqp091.DialConfig(c.url, amqp091.Config{Dial: amqp091.DefaultDial(3 * time.Second)})
	if err != nil {
		return err
	}

	pub, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return err
	}

	consumer, err := conn.Channel()
	if err != nil {
		_ = pub.Close()
		_ = conn.Close()
		return err
	}

	cleanup := func() { _ = consumer.Close(); _ = pub.Close(); _ = conn.Close() }
	if err := pub.Confirm(false); err != nil {
		cleanup()
		return err
	}

	q, err := consumer.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		cleanup()
		return err
	}

	msgs, err := consumer.Consume(q.Name, "", true, true, false, false, nil)
	if err != nil {
		cleanup()
		return err
	}

	c.conn, c.pub, c.consumer = conn, pub, consumer
	c.returns = pub.NotifyReturn(make(chan amqp091.Return, 64))
	c.replyQueue = q.Name
	go c.handleReplies(conn, msgs, c.returns, pub.NotifyClose(make(chan *amqp091.Error, 1)))

	return nil
}

func (c *Client) invalidateLocked() {
	for id, ch := range c.pendingCalls {
		select {
		case ch <- reply{err: errors.ErrServiceUnavailable}:
		default:
		}
		delete(c.pendingCalls, id)
	}

	if c.conn != nil {
		_ = c.conn.Close()
	}

	c.conn, c.pub, c.consumer = nil, nil, nil
	c.replyQueue = ""
}

func (c *Client) handleReplies(conn *amqp091.Connection, msgs <-chan amqp091.Delivery, returns <-chan amqp091.Return, pubClosed <-chan *amqp091.Error) {
	for {
		select {
		case d, ok := <-msgs:
			if !ok {
				c.failConnection(conn)

				return
			}
			c.mu.Lock()
			if c.conn == conn {
				if ch := c.pendingCalls[d.CorrelationId]; ch != nil {
					select {
					case ch <- reply{body: d.Body}:
					default:
					}
					delete(c.pendingCalls, d.CorrelationId)
				}
			}
			c.mu.Unlock()
		case returned, ok := <-returns:
			if !ok {
				c.failConnection(conn)

				return
			}
			c.mu.Lock()
			if c.conn == conn {
				if ch := c.pendingCalls[returned.CorrelationId]; ch != nil {
					select {
					case ch <- reply{err: errors.ErrServiceUnavailable}:
					default:
					}
					delete(c.pendingCalls, returned.CorrelationId)
				}
			}
			c.mu.Unlock()
		case <-pubClosed:
			c.failConnection(conn)

			return
		}
	}
}

func (c *Client) failConnection(conn *amqp091.Connection) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == conn {
		c.invalidateLocked()
	}
}

func (c *Client) Request(ctx context.Context, routingKey string, payload any) ([]byte, error) {
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.ErrInternalServerError
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if ctx.Err() != nil {
		return nil, errors.ErrServiceUnavailable
	}
	select {
	case c.inflight <- struct{}{}:
		defer func() { <-c.inflight }()
	case <-ctx.Done():
		return nil, errors.ErrServiceUnavailable
	}

	corrID := uuid.NewString()
	resChan := make(chan reply, 1)
	c.mu.Lock()

	if err := c.connectLocked(); err != nil {
		c.mu.Unlock()
		return nil, errors.ErrServiceUnavailable
	}

	c.pendingCalls[corrID] = resChan
	pub, replyQueue := c.pub, c.replyQueue
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		delete(c.pendingCalls, corrID)
		c.mu.Unlock()
	}()

	deadline, _ := ctx.Deadline()
	remaining := max(time.Until(deadline), time.Millisecond)
	confirmation, err := pub.PublishWithDeferredConfirmWithContext(ctx, c.exchange, routingKey, true, false, amqp091.Publishing{
		ContentType:   "application/json",
		DeliveryMode:  amqp091.Transient,
		CorrelationId: corrID,
		ReplyTo:       replyQueue,
		Expiration:    fmt.Sprint(remaining.Milliseconds()),
		Body:          requestBody,
	})
	if err == nil {
		var acked bool
		acked, err = confirmation.WaitContext(ctx)
		if err == nil && !acked {
			err = fmt.Errorf("RabbitMQ rejected published request")
		}
	}
	if err != nil {
		return nil, errors.ErrServiceUnavailable
	}

	select {
	case result := <-resChan:
		if result.err != nil {
			return nil, result.err
		}
		var resp struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(result.body, &resp); err != nil {
			return nil, errors.ErrServiceUnavailable
		}
		if resp.Error != "" {
			return nil, errors.FromSlug(resp.Error)
		}
		return result.body, nil
	case <-ctx.Done():
		return nil, errors.ErrServiceUnavailable
	}
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	c.invalidateLocked()
	return nil
}
