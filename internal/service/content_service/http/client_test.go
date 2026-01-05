package http

import (
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
	}{
		{
			name: "successful proxy",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/pdf")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("%PDF-1.4 content"))
			},
			wantErr: nil,
		},
		{
			name: "content service returns 404",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantErr: errors.ErrCVNotFound,
		},
		{
			name: "content service returns 401",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			},
			wantErr: errors.ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer ts.Close()

			c := NewClient(ts.URL)
			w := httptest.NewRecorder()

			err := c.ProxyCVDownload(w, "token", "pl")

			if err != tt.wantErr {
				t.Errorf("ProxyCVDownload() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr == nil && w.Header().Get("Content-Type") != "application/pdf" {
				t.Errorf("ProxyCVDownload() expected Content-Type header to be set")
			}
		})
	}
}
