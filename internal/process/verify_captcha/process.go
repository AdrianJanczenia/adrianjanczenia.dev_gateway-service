package verify_captcha

import (
	"context"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/service/captcha_service/http"
)

type CaptchaClient interface {
	VerifyCaptcha(ctx context.Context, body http.VerifyCaptchaRequest) (*http.VerifyCaptchaResponse, error)
}

type Process struct {
	client CaptchaClient
}

func NewProcess(client CaptchaClient) *Process {
	return &Process{
		client: client,
	}
}

func (p *Process) Process(ctx context.Context, body http.VerifyCaptchaRequest) (*http.VerifyCaptchaResponse, error) {
	return p.client.VerifyCaptcha(ctx, body)
}
