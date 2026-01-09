package get_cv_token

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appErrors "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
)

type mockGetCVTokenProcess struct {
	executeFunc func(ctx context.Context, password, lang, captchaID string) (string, error)
}

func (m *mockGetCVTokenProcess) Process(ctx context.Context, password, lang, captchaID string) (string, error) {
	return m.executeFunc(ctx, password, lang, captchaID)
}

func TestHandler_GetCVToken(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		body        any
		executeFunc func(context.Context, string, string, string) (string, error)
		wantStatus  int
		wantToken   string
	}{
		{
			name:   "success",
			method: http.MethodPost,
			body:   map[string]string{"password": "123", "lang": "pl", "captchaId": "123"},
			executeFunc: func(ctx context.Context, p, l, c string) (string, error) {
				return "tok", nil
			},
			wantStatus: http.StatusOK,
			wantToken:  "tok",
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "invalid input",
			method:     http.MethodPost,
			body:       "invalid-json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "auth error",
			method: http.MethodPost,
			body:   map[string]string{"password": "wrong", "lang": "pl", "captchaId": "123"},
			executeFunc: func(ctx context.Context, p, l, c string) (string, error) {
				return "", appErrors.ErrInvalidPassword
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&mockGetCVTokenProcess{executeFunc: tt.executeFunc})
			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			}
			req := httptest.NewRequest(tt.method, "/api/v1/cv-request", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			h.Handle(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Handle() status = %v, wantStatus %v", w.Code, tt.wantStatus)
			}

			if tt.wantToken != "" {
				var resp map[string]string
				json.Unmarshal(w.Body.Bytes(), &resp)
				if resp["token"] != tt.wantToken {
					t.Errorf("Handle() token = %v, wantToken %v", resp["token"], tt.wantToken)
				}
			}
		})
	}
}
