package get_content

import (
	"encoding/json"
	"net/http"
)

type Process interface {
	Execute(lang string) (string, error)
}

type Handler struct {
	process Process
}

func NewHandler(p Process) *Handler {
	return &Handler{process: p}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "pl"
	}

	contentJSON, err := h.process.Execute(lang)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var content map[string]interface{}
	_ = json.Unmarshal([]byte(contentJSON), &content)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(content)
}
