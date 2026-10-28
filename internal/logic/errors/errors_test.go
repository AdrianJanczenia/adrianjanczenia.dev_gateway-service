package errors

import (
	"net/http"
	"testing"
)

func TestUpstreamErrorContract(t *testing.T) {
	for _, tt := range []struct {
		name   string
		status int
		slug   string
		want   *AppError
	}{
		{"CV expired", http.StatusGone, "error_cv_expired", ErrCVExpired},
		{"CAPTCHA server failure", http.StatusInternalServerError, "error_captcha_server", ErrCaptchaServer},
		{"invalid CAPTCHA answer", http.StatusBadRequest, "error_captcha_invalid", ErrInvalidCaptchaValue},
		{"legacy bad request", http.StatusBadRequest, "error_message", ErrInvalidInput},
		{"legacy service unavailable", http.StatusServiceUnavailable, "error_message", ErrServiceUnavailable},
		{"missing slug", http.StatusNotFound, "", ErrBadGateway},
		{"unknown slug", http.StatusNotFound, "other", ErrBadGateway},
		{"contradictory status", http.StatusNotFound, "error_cv_expired", ErrBadGateway},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromUpstream(tt.status, tt.slug); got != tt.want {
				t.Fatalf("FromUpstream(%d, %q) = %v, want %v", tt.status, tt.slug, got, tt.want)
			}
		})
	}
	if got := FromSlug("error_captcha_not_solved"); got != ErrCaptchaNotSolved {
		t.Fatalf("Rabbit error mapped to %v, want CAPTCHA not solved", got)
	}
}
