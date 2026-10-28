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
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		errors.WriteJSON(w, errors.ErrMethodNotAllowed)
		return
	}

	token := r.URL.Query().Get("token")
	lang := r.URL.Query().Get("lang")

	if token == "" || lang == "" {
		errors.WriteJSON(w, errors.ErrInvalidInput)
		return
	}

	tracked := &responseTracker{ResponseWriter: w}
	if err := h.process.Execute(r.Context(), tracked, token, lang); err != nil && !tracked.started {
		errors.WriteJSON(w, err)
	}
}

type responseTracker struct {
	http.ResponseWriter
	started bool
}

func (w *responseTracker) WriteHeader(status int) {
	w.started = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseTracker) Write(p []byte) (int, error) {
	w.started = true
	return w.ResponseWriter.Write(p)
}
