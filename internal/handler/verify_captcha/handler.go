package verify_captcha

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
	service "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/service/captcha_service/http"
)

type VerifyCaptchaProcess interface {
	Process(ctx context.Context, body service.VerifyCaptchaRequest) (*service.VerifyCaptchaResponse, error)
}

type Handler struct {
	process VerifyCaptchaProcess
}

func NewHandler(process VerifyCaptchaProcess) *Handler {
	return &Handler{
		process: process,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteJSON(w, errors.ErrMethodNotAllowed)
		return
	}

	var body service.VerifyCaptchaRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		errors.WriteJSON(w, errors.ErrInvalidInput)
		return
	}

	resp, err := h.process.Process(r.Context(), body)
	if err != nil {
		errors.WriteJSON(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
