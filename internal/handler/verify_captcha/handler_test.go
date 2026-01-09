package verify_captcha

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

type mockVerifyCaptchaProcess struct {
	processFunc func(ctx context.Context, body service.VerifyCaptchaRequest) (*service.VerifyCaptchaResponse, error)
}

func (m *mockVerifyCaptchaProcess) Process(ctx context.Context, body service.VerifyCaptchaRequest) (*service.VerifyCaptchaResponse, error) {
	return m.processFunc(ctx, body)
}

func TestHandler_VerifyCaptcha(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		body        interface{}
		processFunc func(ctx context.Context, body service.VerifyCaptchaRequest) (*service.VerifyCaptchaResponse, error)
		wantStatus  int
	}{
		{
			name:   "success",
			method: http.MethodPost,
			body:   service.VerifyCaptchaRequest{CaptchaId: "id", CaptchaValue: "123"},
			processFunc: func(ctx context.Context, body service.VerifyCaptchaRequest) (*service.VerifyCaptchaResponse, error) {
				return &service.VerifyCaptchaResponse{CaptchaId: "id"}, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "process error",
			method: http.MethodPost,
			body:   service.VerifyCaptchaRequest{CaptchaId: "id", CaptchaValue: "123"},
			processFunc: func(ctx context.Context, body service.VerifyCaptchaRequest) (*service.VerifyCaptchaResponse, error) {
				return nil, errors.New("err")
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&mockVerifyCaptchaProcess{processFunc: tt.processFunc})
			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, "/api/v1/captcha-verify", bytes.NewBuffer(b))
			w := httptest.NewRecorder()

			h.Handle(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Handle() status = %v, wantStatus %v", w.Code, tt.wantStatus)
			}
		})
	}
}
