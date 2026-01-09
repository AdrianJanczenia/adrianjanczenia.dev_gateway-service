package get_pow

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	service "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/service/captcha_service/http"
)

type mockGetPowProcess struct {
	processFunc func(ctx context.Context) (*service.GetPowResponse, error)
}

func (m *mockGetPowProcess) Process(ctx context.Context) (*service.GetPowResponse, error) {
	return m.processFunc(ctx)
}

func TestHandler_GetPow(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		processFunc func(ctx context.Context) (*service.GetPowResponse, error)
		wantStatus  int
	}{
		{
			name:   "success",
			method: http.MethodGet,
			processFunc: func(ctx context.Context) (*service.GetPowResponse, error) {
				return &service.GetPowResponse{Seed: "s", Signature: "sig"}, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "method not allowed",
			method:      http.MethodPost,
			processFunc: nil,
			wantStatus:  http.StatusMethodNotAllowed,
		},
		{
			name:   "process error",
			method: http.MethodGet,
			processFunc: func(ctx context.Context) (*service.GetPowResponse, error) {
				return nil, errors.New("err")
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&mockGetPowProcess{processFunc: tt.processFunc})
			req := httptest.NewRequest(tt.method, "/api/v1/pow", nil)
			w := httptest.NewRecorder()

			h.Handle(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Handle() status = %v, wantStatus %v", w.Code, tt.wantStatus)
			}
		})
	}
}
