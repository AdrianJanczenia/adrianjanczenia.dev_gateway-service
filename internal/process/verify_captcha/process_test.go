package verify_captcha

import (
	"context"
	"errors"
	"testing"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/service/captcha_service/http"
)

type mockCaptchaClient struct {
	verifyCaptchaFunc func(ctx context.Context, body http.VerifyCaptchaRequest) (*http.VerifyCaptchaResponse, error)
}

func (m *mockCaptchaClient) VerifyCaptcha(ctx context.Context, body http.VerifyCaptchaRequest) (*http.VerifyCaptchaResponse, error) {
	return m.verifyCaptchaFunc(ctx, body)
}

func TestProcess_VerifyCaptcha(t *testing.T) {
	tests := []struct {
		name              string
		verifyCaptchaFunc func(ctx context.Context, body http.VerifyCaptchaRequest) (*http.VerifyCaptchaResponse, error)
		want              *http.VerifyCaptchaResponse
		wantErr           bool
	}{
		{
			name: "success",
			verifyCaptchaFunc: func(ctx context.Context, body http.VerifyCaptchaRequest) (*http.VerifyCaptchaResponse, error) {
				return &http.VerifyCaptchaResponse{CaptchaId: "id"}, nil
			},
			want:    &http.VerifyCaptchaResponse{CaptchaId: "id"},
			wantErr: false,
		},
		{
			name: "error",
			verifyCaptchaFunc: func(ctx context.Context, body http.VerifyCaptchaRequest) (*http.VerifyCaptchaResponse, error) {
				return nil, errors.New("err")
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockCaptchaClient{verifyCaptchaFunc: tt.verifyCaptchaFunc}
			p := NewProcess(m)
			got, err := p.Process(context.Background(), http.VerifyCaptchaRequest{})
			if (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.want != nil && got.CaptchaId != tt.want.CaptchaId {
				t.Errorf("Process() got = %v, want %v", got, tt.want)
			}
		})
	}
}
