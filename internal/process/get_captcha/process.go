package get_captcha

import (
	"context"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/service/captcha_service/http"
)

type CaptchaClient interface {
	GetCaptcha(ctx context.Context, body http.GetCaptchaRequest) (*http.GetCaptchaResponse, error)
}

type Process struct {
	client CaptchaClient
}

func NewProcess(client CaptchaClient) *Process {
	return &Process{
		client: client,
	}
}

func (p *Process) Process(ctx context.Context, body http.GetCaptchaRequest) (*http.GetCaptchaResponse, error) {
	return p.client.GetCaptcha(ctx, body)
}
