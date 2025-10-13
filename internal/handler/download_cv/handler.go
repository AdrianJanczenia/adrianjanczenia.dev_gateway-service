package download_cv

import (
	"net/http"
)

type Process interface {
	Execute(w http.ResponseWriter, r *http.Request) error
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

	if err := h.process.Execute(w, r); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
