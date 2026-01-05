package download_cv

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockContentServiceClient struct {
	proxyCVDownloadFunc func(ctx context.Context, w http.ResponseWriter, token string, lang string) error
}

func (m *mockContentServiceClient) ProxyCVDownload(ctx context.Context, w http.ResponseWriter, token string, lang string) error {
	return m.proxyCVDownloadFunc(ctx, w, token, lang)
}

func TestProcess_DownloadCV(t *testing.T) {
	tests := []struct {
		name                string
		proxyCVDownloadFunc func(ctx context.Context, w http.ResponseWriter, token string, lang string) error
		wantErr             bool
	}{
		{
			name: "success",
			proxyCVDownloadFunc: func(ctx context.Context, w http.ResponseWriter, token string, lang string) error {
				w.WriteHeader(http.StatusOK)
				return nil
			},
			wantErr: false,
		},
		{
			name: "error from client",
			proxyCVDownloadFunc: func(ctx context.Context, w http.ResponseWriter, token string, lang string) error {
				return errors.New("download failed")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockContentServiceClient{proxyCVDownloadFunc: tt.proxyCVDownloadFunc}
			p := NewProcess(m)
			w := httptest.NewRecorder()
			err := p.Execute(context.Background(), w, "test-token", "pl")
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
