package get_pow

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
	service "github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/service/captcha_service/http"
)

type GetPowProcess interface {
	Process(ctx context.Context) (*service.GetPowResponse, error)
}

type Handler struct {
	process GetPowProcess
}

func NewHandler(process GetPowProcess) *Handler {
	return &Handler{
		process: process,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteJSON(w, errors.ErrMethodNotAllowed)
		return
	}

	resp, err := h.process.Process(r.Context())
	if err != nil {
		errors.WriteJSON(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
