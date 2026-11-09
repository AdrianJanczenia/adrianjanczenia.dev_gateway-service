package errors

import (
	"encoding/json"
	"errors"
	"net/http"
)

type AppError struct {
	HTTPStatus int
	Slug       string
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Slug
}

func (e *AppError) Unwrap() error {
	return e.Err
}

var (
	ErrInternalServerError = &AppError{HTTPStatus: http.StatusInternalServerError, Slug: "error_internal"}
	ErrCVServer            = &AppError{HTTPStatus: http.StatusInternalServerError, Slug: "error_cv_server"}
	ErrServiceUnavailable  = &AppError{HTTPStatus: http.StatusServiceUnavailable, Slug: "error_service_unavailable"}
	ErrBadGateway          = &AppError{HTTPStatus: http.StatusBadGateway, Slug: "error_bad_gateway"}
	ErrMethodNotAllowed    = &AppError{HTTPStatus: http.StatusMethodNotAllowed, Slug: "error_method_not_allowed"}
	ErrInvalidInput        = &AppError{HTTPStatus: http.StatusBadRequest, Slug: "error_invalid_input"}
	ErrUnsupportedLanguage = &AppError{HTTPStatus: http.StatusBadRequest, Slug: "error_unsupported_language"}
	ErrInvalidCode         = &AppError{HTTPStatus: http.StatusUnauthorized, Slug: "error_cv_auth"}
	ErrCVNotFound          = &AppError{HTTPStatus: http.StatusNotFound, Slug: "error_cv_not_found"}
	ErrCVExpired           = &AppError{HTTPStatus: http.StatusGone, Slug: "error_cv_expired"}
	ErrCVCodeExpired       = &AppError{HTTPStatus: http.StatusGone, Slug: "error_cv_code_expired"}
	ErrCVCodeInactive      = &AppError{HTTPStatus: http.StatusGone, Slug: "error_cv_code_inactive"}
	ErrCVCodeLimit         = &AppError{HTTPStatus: http.StatusGone, Slug: "error_cv_code_limit"}
	ErrContentNotFound     = &AppError{HTTPStatus: http.StatusNotFound, Slug: "error_content_not_found"}
	ErrInvalidSignature    = &AppError{HTTPStatus: http.StatusForbidden, Slug: "error_pow_signature"}
	ErrSeedAlreadyUsed     = &AppError{HTTPStatus: http.StatusConflict, Slug: "error_pow_double_spend"}
	ErrPowExpired          = &AppError{HTTPStatus: http.StatusGone, Slug: "error_pow_expired"}
	ErrInsufficientWork    = &AppError{HTTPStatus: http.StatusBadRequest, Slug: "error_pow_work"}
	ErrCaptchaNotFound     = &AppError{HTTPStatus: http.StatusNotFound, Slug: "error_captcha_not_found"}
	ErrInvalidCaptchaValue = &AppError{HTTPStatus: http.StatusBadRequest, Slug: "error_captcha_invalid"}
	ErrCaptchaNotSolved    = &AppError{HTTPStatus: http.StatusForbidden, Slug: "error_captcha_not_solved"}
	ErrNoTriesLeft         = &AppError{HTTPStatus: http.StatusGone, Slug: "error_captcha_expired"}
	ErrCaptchaServer       = &AppError{HTTPStatus: http.StatusInternalServerError, Slug: "error_captcha_server"}
)

func FromSlug(slug string) *AppError {
	switch slug {
	case "error_cv_auth":
		return ErrInvalidCode
	case "error_cv_expired":
		return ErrCVExpired
	case "error_cv_code_expired":
		return ErrCVCodeExpired
	case "error_cv_code_inactive":
		return ErrCVCodeInactive
	case "error_cv_code_limit":
		return ErrCVCodeLimit
	case "error_cv_not_found":
		return ErrCVNotFound
	case "error_cv_server":
		return ErrCVServer
	case "error_internal":
		return ErrInternalServerError
	case "error_captcha_server":
		return ErrCaptchaServer
	case "error_pow_signature":
		return ErrInvalidSignature
	case "error_pow_double_spend":
		return ErrSeedAlreadyUsed
	case "error_pow_expired":
		return ErrPowExpired
	case "error_pow_work":
		return ErrInsufficientWork
	case "error_captcha_not_found":
		return ErrCaptchaNotFound
	case "error_captcha_invalid":
		return ErrInvalidCaptchaValue
	case "error_captcha_not_solved":
		return ErrCaptchaNotSolved
	case "error_captcha_expired":
		return ErrNoTriesLeft
	case "error_message":
		return ErrServiceUnavailable
	case "error_service_unavailable":
		return ErrServiceUnavailable
	case "error_bad_gateway":
		return ErrBadGateway
	case "error_method_not_allowed":
		return ErrMethodNotAllowed
	case "error_invalid_input":
		return ErrInvalidInput
	case "error_unsupported_language":
		return ErrUnsupportedLanguage
	case "error_content_not_found":
		return ErrContentNotFound
	default:
		return ErrInternalServerError
	}
}

func FromUpstream(status int, slug string) *AppError {
	if slug == "error_message" {
		switch status {
		case http.StatusBadRequest:
			return ErrInvalidInput
		case http.StatusMethodNotAllowed:
			return ErrMethodNotAllowed
		case http.StatusNotFound:
			return ErrContentNotFound
		case http.StatusServiceUnavailable:
			return ErrServiceUnavailable
		}
	}
	appErr := FromSlug(slug)
	if slug == appErr.Slug && status == appErr.HTTPStatus {
		return appErr
	}
	return ErrBadGateway
}

func WriteJSON(w http.ResponseWriter, err error) {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		appErr = ErrInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatus)
	json.NewEncoder(w).Encode(map[string]string{
		"error": appErr.Slug,
	})
}
