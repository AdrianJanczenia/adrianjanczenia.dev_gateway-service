package get_content

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockGetContentProcess struct {
	processFunc func(ctx context.Context, lang string) ([]byte, error)
}

func (m *mockGetContentProcess) Process(ctx context.Context, lang string) ([]byte, error) {
	return m.processFunc(ctx, lang)
}

func TestHandler_GetContent(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		processFunc func(ctx context.Context, lang string) ([]byte, error)
		wantStatus  int
		wantBody    string
	}{
		{
			name: "success with language",
			url:  "/api/v1/content?lang=pl",
			processFunc: func(ctx context.Context, lang string) ([]byte, error) {
				if lang != "pl" {
					return nil, errors.New("wrong lang")
				}
				return []byte(`{"key":"value"}`), nil
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"key":"value"}`,
		},
		{
			name: "success default language",
			url:  "/api/v1/content",
			processFunc: func(ctx context.Context, lang string) ([]byte, error) {
				if lang != "en" {
					return nil, errors.New("wrong lang")
				}
				return []byte(`{"hello":"world"}`), nil
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"hello":"world"}`,
		},
		{
			name: "process error",
			url:  "/api/v1/content?lang=pl",
			processFunc: func(ctx context.Context, lang string) ([]byte, error) {
				return nil, errors.New("internal error")
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&mockGetContentProcess{processFunc: tt.processFunc})
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			h.Handle(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Handle() status = %v, wantStatus %v", w.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusOK && w.Body.String() != tt.wantBody {
				t.Errorf("Handle() body = %v, wantBody %v", w.Body.String(), tt.wantBody)
			}
		})
	}
}
