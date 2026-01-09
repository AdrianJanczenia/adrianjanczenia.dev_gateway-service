package get_pow

import (
	"context"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/service/captcha_service/http"
)

type CaptchaClient interface {
	GetPow(ctx context.Context) (*http.GetPowResponse, error)
}

type Process struct {
	client CaptchaClient
}

func NewProcess(client CaptchaClient) *Process {
	return &Process{
		client: client,
	}
}

func (p *Process) Process(ctx context.Context) (*http.GetPowResponse, error) {
	return p.client.GetPow(ctx)
}
