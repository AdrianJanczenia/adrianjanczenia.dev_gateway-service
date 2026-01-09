package get_cv_token

import (
	"context"
	"errors"
	"testing"
)

type mockRabbitMQClient struct {
	requestFunc func(ctx context.Context, routingKey string, payload any) (body []byte, err error)
}

func (m *mockRabbitMQClient) Request(ctx context.Context, routingKey string, payload any) (body []byte, err error) {
	return m.requestFunc(ctx, routingKey, payload)
}

func TestProcess_GetCVToken(t *testing.T) {
	tests := []struct {
		name        string
		requestFunc func(context.Context, string, any) ([]byte, error)
		want        string
		wantErr     bool
	}{
		{
			name: "successful token retrieval",
			requestFunc: func(ctx context.Context, rk string, p any) ([]byte, error) {
				return []byte(`{"token":"secret-token"}`), nil
			},
			want:    "secret-token",
			wantErr: false,
		},
		{
			name: "rabbitmq error",
			requestFunc: func(ctx context.Context, rk string, p any) ([]byte, error) {
				return nil, errors.New("amqp error")
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "malformed response",
			requestFunc: func(ctx context.Context, rk string, p any) ([]byte, error) {
				return []byte(`{invalid}`), nil
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockRabbitMQClient{requestFunc: tt.requestFunc}
			p := NewProcess(m, "key")
			got, err := p.Process(context.Background(), "pass", "pl", "123")
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Execute() got = %v, want %v", got, tt.want)
			}
		})
	}
}
