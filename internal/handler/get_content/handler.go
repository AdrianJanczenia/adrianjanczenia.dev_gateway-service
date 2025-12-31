package get_content

import (
	"context"
	"net/http"
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
		http.Error(w, "Failed to get content", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(content)
}
