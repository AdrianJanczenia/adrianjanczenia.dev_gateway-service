package get_pow

import (
	"context"
	"errors"
	"testing"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/service/captcha_service/http"
)

type mockCaptchaClient struct {
	getPowFunc func(ctx context.Context) (*http.GetPowResponse, error)
}

func (m *mockCaptchaClient) GetPow(ctx context.Context) (*http.GetPowResponse, error) {
	return m.getPowFunc(ctx)
}

func TestProcess_GetPow(t *testing.T) {
	tests := []struct {
		name       string
		getPowFunc func(ctx context.Context) (*http.GetPowResponse, error)
		want       *http.GetPowResponse
		wantErr    bool
	}{
		{
			name: "success",
			getPowFunc: func(ctx context.Context) (*http.GetPowResponse, error) {
				return &http.GetPowResponse{Seed: "s", Signature: "sig"}, nil
			},
			want:    &http.GetPowResponse{Seed: "s", Signature: "sig"},
			wantErr: false,
		},
		{
			name: "error",
			getPowFunc: func(ctx context.Context) (*http.GetPowResponse, error) {
				return nil, errors.New("err")
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockCaptchaClient{getPowFunc: tt.getPowFunc}
			p := NewProcess(m)
			got, err := p.Process(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.want != nil && (got.Seed != tt.want.Seed || got.Signature != tt.want.Signature) {
				t.Errorf("Process() got = %v, want %v", got, tt.want)
			}
		})
	}
}
