package get_cv_token

import (
	"encoding/json"
	"net/http"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
)

type GetCVTokenProcess interface {
	Execute(password, lang string) (string, error)
}

type Handler struct {
	process GetCVTokenProcess
}

func NewHandler(p GetCVTokenProcess) *Handler {
	return &Handler{process: p}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errors.WriteJSON(w, errors.ErrMethodNotAllowed)
		return
	}

	var payload struct {
		Password string `json:"password"`
		Lang     string `json:"lang"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		errors.WriteJSON(w, errors.ErrInvalidInput)
		return
	}

	cvToken, err := h.process.Execute(payload.Password, payload.Lang)
	if err != nil {
		errors.WriteJSON(w, err)
		return
	}

	response := map[string]string{"token": cvToken}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
