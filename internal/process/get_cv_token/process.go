package get_cv_token

import (
	"context"
	"encoding/json"
)

type RabbitMQClient interface {
	Request(ctx context.Context, routingKey string, payload any) (body []byte, err error)
}

type Process struct {
	rabbitClient RabbitMQClient
	routingKey   string
}

func NewProcess(client RabbitMQClient, routingKey string) *Process {
	return &Process{
		rabbitClient: client,
		routingKey:   routingKey,
	}
}

func (p *Process) Execute(ctx context.Context, password, lang string) (string, error) {
	payload := map[string]string{"password": password, "lang": lang}

	responseBody, err := p.rabbitClient.Request(ctx, p.routingKey, payload)
	if err != nil {
		return "", err
	}

	var resp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		return "", err
	}

	return resp.Token, nil
}
