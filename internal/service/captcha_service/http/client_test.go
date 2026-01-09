package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
)

func TestClient_GetPow(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		want           *GetPowResponse
		wantErr        error
	}{
		{
			name: "success",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{"seed": "test_seed", "signature": "test_sig"})
			},
			want:    &GetPowResponse{Seed: "test_seed", Signature: "test_sig"},
			wantErr: nil,
		},
		{
			name: "error from service",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{"error": "error_pow_signature"})
			},
			want:    nil,
			wantErr: errors.ErrInvalidSignature,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer ts.Close()

			c := NewClient(&http.Client{}, ts.URL)
			got, err := c.GetPow(context.Background())

			if err != tt.wantErr {
				t.Errorf("GetPow() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.want != nil && (got.Seed != tt.want.Seed || got.Signature != tt.want.Signature) {
				t.Errorf("GetPow() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetCaptcha(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		want           *GetCaptchaResponse
		wantErr        error
	}{
		{
			name: "success",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{"captchaId": "id123", "captchaImg": "img_b64"})
			},
			want:    &GetCaptchaResponse{CaptchaId: "id123", CaptchaImg: "img_b64"},
			wantErr: nil,
		},
		{
			name: "invalid work error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "error_pow_work"})
			},
			want:    nil,
			wantErr: errors.ErrInsufficientWork,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer ts.Close()

			c := NewClient(&http.Client{}, ts.URL)
			got, err := c.GetCaptcha(context.Background(), GetCaptchaRequest{})

			if err != tt.wantErr {
				t.Errorf("GetCaptcha() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.want != nil && (got.CaptchaId != tt.want.CaptchaId || got.CaptchaImg != tt.want.CaptchaImg) {
				t.Errorf("GetCaptcha() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_VerifyCaptcha(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		want           *VerifyCaptchaResponse
		wantErr        error
	}{
		{
			name: "success",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{"captchaId": "id123"})
			},
			want:    &VerifyCaptchaResponse{CaptchaId: "id123"},
			wantErr: nil,
		},
		{
			name: "not found error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "error_captcha_not_found"})
			},
			want:    nil,
			wantErr: errors.ErrCaptchaNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer ts.Close()

			c := NewClient(&http.Client{}, ts.URL)
			got, err := c.VerifyCaptcha(context.Background(), VerifyCaptchaRequest{})

			if err != tt.wantErr {
				t.Errorf("VerifyCaptcha() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.want != nil && got.CaptchaId != tt.want.CaptchaId {
				t.Errorf("VerifyCaptcha() got = %v, want %v", got, tt.want)
			}
		})
	}
}
