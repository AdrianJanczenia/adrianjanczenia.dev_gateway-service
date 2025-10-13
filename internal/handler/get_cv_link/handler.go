package get_cv_link

import (
	"encoding/json"
	"net/http"
)

type Process interface {
	Execute(password, lang string) (string, error)
}

type Handler struct {
	process Process
}

func NewHandler(p Process) *Handler {
	return &Handler{process: p}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Password string `json:"password"`
		Lang     string `json:"lang"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	downloadURL, err := h.process.Execute(payload.Password, payload.Lang)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response := map[string]string{"downloadUrl": downloadURL}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
