package get_captcha

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	service "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/service/captcha_service/http"
)

type mockGetCaptchaProcess struct {
	processFunc func(ctx context.Context, body service.GetCaptchaRequest) (*service.GetCaptchaResponse, error)
}

func (m *mockGetCaptchaProcess) Process(ctx context.Context, body service.GetCaptchaRequest) (*service.GetCaptchaResponse, error) {
	return m.processFunc(ctx, body)
}

func TestHandler_GetCaptcha(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		body        interface{}
		processFunc func(ctx context.Context, body service.GetCaptchaRequest) (*service.GetCaptchaResponse, error)
		wantStatus  int
	}{
		{
			name:   "success",
			method: http.MethodPost,
			body:   service.GetCaptchaRequest{Seed: "s"},
			processFunc: func(ctx context.Context, body service.GetCaptchaRequest) (*service.GetCaptchaResponse, error) {
				return &service.GetCaptchaResponse{CaptchaId: "id"}, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "invalid input",
			method:      http.MethodPost,
			body:        "invalid",
			processFunc: nil,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:   "process error",
			method: http.MethodPost,
			body:   service.GetCaptchaRequest{Seed: "s"},
			processFunc: func(ctx context.Context, body service.GetCaptchaRequest) (*service.GetCaptchaResponse, error) {
				return nil, errors.New("err")
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&mockGetCaptchaProcess{processFunc: tt.processFunc})
			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, "/api/v1/captcha", bytes.NewBuffer(b))
			w := httptest.NewRecorder()

			h.Handle(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Handle() status = %v, wantStatus %v", w.Code, tt.wantStatus)
			}
		})
	}
}
