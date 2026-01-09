package get_captcha

import (
	"context"
	"errors"
	"testing"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/service/captcha_service/http"
)

type mockCaptchaClient struct {
	getCaptchaFunc func(ctx context.Context, body http.GetCaptchaRequest) (*http.GetCaptchaResponse, error)
}

func (m *mockCaptchaClient) GetCaptcha(ctx context.Context, body http.GetCaptchaRequest) (*http.GetCaptchaResponse, error) {
	return m.getCaptchaFunc(ctx, body)
}

func TestProcess_GetCaptcha(t *testing.T) {
	tests := []struct {
		name           string
		getCaptchaFunc func(ctx context.Context, body http.GetCaptchaRequest) (*http.GetCaptchaResponse, error)
		want           *http.GetCaptchaResponse
		wantErr        bool
	}{
		{
			name: "success",
			getCaptchaFunc: func(ctx context.Context, body http.GetCaptchaRequest) (*http.GetCaptchaResponse, error) {
				return &http.GetCaptchaResponse{CaptchaId: "id", CaptchaImg: "img"}, nil
			},
			want:    &http.GetCaptchaResponse{CaptchaId: "id", CaptchaImg: "img"},
			wantErr: false,
		},
		{
			name: "error",
			getCaptchaFunc: func(ctx context.Context, body http.GetCaptchaRequest) (*http.GetCaptchaResponse, error) {
				return nil, errors.New("err")
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockCaptchaClient{getCaptchaFunc: tt.getCaptchaFunc}
			p := NewProcess(m)
			got, err := p.Process(context.Background(), http.GetCaptchaRequest{})
			if (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.want != nil && (got.CaptchaId != tt.want.CaptchaId || got.CaptchaImg != tt.want.CaptchaImg) {
				t.Errorf("Process() got = %v, want %v", got, tt.want)
			}
		})
	}
}
