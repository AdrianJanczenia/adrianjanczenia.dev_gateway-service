package download_cv

import (
	"context"
	"net/http"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
)

type DownloadCVProcess interface {
	Execute(ctx context.Context, w http.ResponseWriter, token string, lang string) error
}

type Handler struct {
	process DownloadCVProcess
}

func NewHandler(p DownloadCVProcess) *Handler {
	return &Handler{process: p}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.WriteJSON(w, errors.ErrMethodNotAllowed)
		return
	}

	token := r.URL.Query().Get("token")
	lang := r.URL.Query().Get("lang")

	if token == "" || lang == "" {
		errors.WriteJSON(w, errors.ErrInvalidInput)
		return
	}

	if err := h.process.Execute(r.Context(), w, token, lang); err != nil {
		errors.WriteJSON(w, err)
		return
	}
}
