package rabbitmq

import (
	"context"
	"testing"
	"time"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
	"github.com/rabbitmq/amqp091-go"
)

func TestRepliesRouteByCorrelationID(t *testing.T) {
	conn := &amqp091.Connection{}
	returnedRequest := make(chan reply, 1)
	answeredRequest := make(chan reply, 1)
	client := &Client{
		conn: conn,
		pendingCalls: map[string]chan reply{
			"returned": returnedRequest,
			"answered": answeredRequest,
		},
	}
	deliveries := make(chan amqp091.Delivery)
	returns := make(chan amqp091.Return)
	closed := make(chan *amqp091.Error)
	done := make(chan struct{})
	go func() {
		client.handleReplies(conn, deliveries, returns, closed)
		close(done)
	}()
	t.Cleanup(func() {
		client.mu.Lock()
		client.conn = nil
		client.mu.Unlock()
		close(closed)
		<-done
	})

	returns <- amqp091.Return{CorrelationId: "returned"}
	select {
	case result := <-returnedRequest:
		if result.err != errors.ErrServiceUnavailable {
			t.Fatalf("unroutable request error = %v", result.err)
		}
	case <-time.After(time.Second):
		t.Fatal("unroutable request was not notified")
	}
	select {
	case result := <-answeredRequest:
		t.Fatalf("unrelated request was notified: %+v", result)
	default:
	}

	deliveries <- amqp091.Delivery{CorrelationId: "answered", Body: []byte(`{"token":"abc"}`)}
	select {
	case result := <-answeredRequest:
		if string(result.body) != `{"token":"abc"}` || result.err != nil {
			t.Fatalf("reply = %+v", result)
		}
	case <-time.After(time.Second):
		t.Fatal("answered request was not notified")
	}
	client.mu.Lock()
	remaining := len(client.pendingCalls)
	client.mu.Unlock()
	if remaining != 0 {
		t.Fatalf("%d requests remain pending", remaining)
	}
}

func TestCloseFailsPendingRequests(t *testing.T) {
	result := make(chan reply, 1)
	client := &Client{pendingCalls: map[string]chan reply{"request": result}}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	if got := <-result; got.err != errors.ErrServiceUnavailable {
		t.Fatalf("pending request error = %v", got.err)
	}
	if len(client.pendingCalls) != 0 || !client.closed {
		t.Fatalf("client was not closed")
	}
}

func TestRequestRespectsDeadlineAtCapacity(t *testing.T) {
	client := &Client{inflight: make(chan struct{}, 1)}
	client.inflight <- struct{}{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err := client.Request(ctx, "cv.requests", map[string]string{"id": "session"})
	if err != errors.ErrServiceUnavailable {
		t.Fatalf("Request error = %v", err)
	}
}
