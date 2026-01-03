package download_cv

import (
	"log"
	"net/http"
)

type Process interface {
	Execute(w http.ResponseWriter, token string, lang string) error
}

type Handler struct {
	process Process
}

func NewHandler(p Process) *Handler {
	return &Handler{process: p}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.URL.Query().Get("token")
	lang := r.URL.Query().Get("lang")

	if token == "" || lang == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.process.Execute(w, token, lang); err != nil {
		log.Printf("ERROR: CV download proxy failed: %v", err)
		return
	}
}
