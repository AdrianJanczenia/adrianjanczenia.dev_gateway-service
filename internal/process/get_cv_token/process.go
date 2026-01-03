package get_cv_token

import (
	"encoding/json"
	"errors"
)

type RabbitMQClient interface {
	Request(routingKey string, payload any) (body []byte, err error)
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

func (p *Process) Execute(password, lang string) (string, error) {
	payload := map[string]string{"password": password, "lang": lang}

	responseBody, err := p.rabbitClient.Request(p.routingKey, payload)
	if err != nil {
		return "", err
	}

	var resp struct {
		Token string `json:"token"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(responseBody, &resp); err != nil {
		return "", err
	}

	if resp.Error != "" {
		return "", errors.New(resp.Error)
	}

	return resp.Token, nil
}
