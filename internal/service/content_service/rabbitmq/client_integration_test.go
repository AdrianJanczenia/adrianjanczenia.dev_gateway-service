//go:build integration

package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	appErrors "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

func TestClientRPCWithRabbitMQ(t *testing.T) {
	rabbitURL := os.Getenv("RABBIT_TEST_URL")
	if rabbitURL == "" {
		t.Skip("RABBIT_TEST_URL is not set")
	}
	connection, err := amqp091.Dial(rabbitURL)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	channel, err := connection.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer channel.Close()
	exchange := "audit.gateway." + uuid.NewString()
	queue := exchange
	const routingKey = "cv.request"
	if err := channel.ExchangeDeclare(exchange, "direct", false, false, false, false, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := channel.QueueDeclare(queue, false, false, false, false, nil); err != nil {
		t.Fatal(err)
	}
	defer func() {
		channel.QueueDelete(queue, false, false, false)
		channel.ExchangeDelete(exchange, false, false)
	}()
	if err := channel.QueueBind(queue, routingKey, exchange, false, nil); err != nil {
		t.Fatal(err)
	}
	const consumerTag = "audit-worker"
	deliveries, err := channel.Consume(queue, consumerTag, true, false, false, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	workerErrors := make(chan error, 1)
	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		for delivery := range deliveries {
			var request struct {
				ID int `json:"id"`
			}
			if err := json.Unmarshal(delivery.Body, &request); err != nil {
				workerErrors <- err
				return
			}
			body, err := json.Marshal(map[string]string{"token": fmt.Sprintf("token-%d", request.ID)})
			if err != nil {
				workerErrors <- err
				return
			}
			if err := channel.PublishWithContext(context.Background(), "", delivery.ReplyTo, true, false, amqp091.Publishing{
				ContentType: "application/json", CorrelationId: delivery.CorrelationId, Body: body,
			}); err != nil {
				workerErrors <- err
				return
			}
		}
	}()
	defer func() {
		if err := channel.Cancel(consumerTag, false); err != nil {
			t.Errorf("cancel test consumer: %v", err)
		}
		<-workerDone
	}()
	client, err := NewClient(rabbitURL, exchange)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	const requests = 16
	var wg sync.WaitGroup
	results := make(chan error, requests)
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			body, err := client.Request(context.Background(), routingKey, map[string]int{"id": id})
			if err != nil {
				results <- err
				return
			}
			var response struct {
				Token string `json:"token"`
			}
			if err := json.Unmarshal(body, &response); err != nil || response.Token != fmt.Sprintf("token-%d", id) {
				results <- fmt.Errorf("request %d received %q: %v", id, response.Token, err)
			}
		}(i)
	}
	wg.Wait()
	close(results)
	for err := range results {
		t.Error(err)
	}
	select {
	case err := <-workerErrors:
		t.Fatal(err)
	default:
	}

	start := time.Now()
	_, err = client.Request(context.Background(), "not.bound", map[string]int{"id": 99})
	if err != appErrors.ErrServiceUnavailable || time.Since(start) > 3*time.Second {
		t.Fatalf("unroutable request: err=%v duration=%v", err, time.Since(start))
	}
	if err := client.conn.Close(); err != nil {
		t.Fatal(err)
	}
	body, err := client.Request(context.Background(), routingKey, map[string]int{"id": 100})
	if err != nil {
		t.Fatalf("request after reconnect: %v", err)
	}
	if string(body) != `{"token":"token-100"}` {
		t.Fatalf("request after reconnect returned %s", body)
	}
}
