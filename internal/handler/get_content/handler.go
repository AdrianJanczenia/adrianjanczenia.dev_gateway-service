package get_content

import (
	"context"
	"net/http"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
)

type GetContentProcess interface {
	Process(ctx context.Context, lang string) ([]byte, error)
}

type Handler struct {
	process GetContentProcess
}

func NewHandler(process GetContentProcess) *Handler {
	return &Handler{process: process}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}

	content, err := h.process.Process(r.Context(), lang)
	if err != nil {
		errors.WriteJSON(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(content)
}
