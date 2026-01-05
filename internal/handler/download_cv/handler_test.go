package download_cv

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockDownloadCVProcess struct {
	executeFunc func(ctx context.Context, w http.ResponseWriter, token string, lang string) error
}

func (m *mockDownloadCVProcess) Execute(ctx context.Context, w http.ResponseWriter, token string, lang string) error {
	return m.executeFunc(ctx, w, token, lang)
}

func TestHandler_DownloadCV(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		url         string
		executeFunc func(context.Context, http.ResponseWriter, string, string) error
		wantStatus  int
	}{
		{
			name:   "success",
			method: http.MethodGet,
			url:    "/api/v1/download/cv?token=abc&lang=pl",
			executeFunc: func(ctx context.Context, w http.ResponseWriter, t, l string) error {
				w.WriteHeader(http.StatusOK)
				return nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing params",
			method:     http.MethodGet,
			url:        "/api/v1/download/cv?token=abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "internal error",
			method: http.MethodGet,
			url:    "/api/v1/download/cv?token=abc&lang=pl",
			executeFunc: func(ctx context.Context, w http.ResponseWriter, t, l string) error {
				return errors.New("fail")
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&mockDownloadCVProcess{executeFunc: tt.executeFunc})
			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			h.Handle(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Handle() status = %v, wantStatus %v", w.Code, tt.wantStatus)
			}
		})
	}
}
