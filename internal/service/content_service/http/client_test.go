package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
)

func TestHTTPClient_ProxyCVDownload(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		wantErr        error
		checkHeaders   bool
	}{
		{
			name: "successful proxy and allow-list check",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/pdf")
				w.Header().Set("X-Internal-Server", "secret-id")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("%PDF-1.4 content"))
			},
			wantErr:      nil,
			checkHeaders: true,
		},
		{
			name: "content service returns 404",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantErr: errors.ErrCVNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer ts.Close()

			c := NewClient(ts.URL)
			w := httptest.NewRecorder()

			err := c.ProxyCVDownload(context.Background(), w, "token", "pl")

			if err != tt.wantErr {
				t.Errorf("ProxyCVDownload() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.checkHeaders {
				if w.Header().Get("Content-Type") != "application/pdf" {
					t.Errorf("ProxyCVDownload() expected Content-Type header to be preserved")
				}
				if w.Header().Get("X-Internal-Server") != "" {
					t.Errorf("ProxyCVDownload() expected X-Internal-Server header to be filtered out")
				}
			}
		})
	}
}
