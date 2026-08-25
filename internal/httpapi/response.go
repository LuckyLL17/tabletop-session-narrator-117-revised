package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"t117/internal/domain"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": value})
}
func writeOK(w http.ResponseWriter, value any)      { writeJSON(w, http.StatusOK, value) }
func writeCreated(w http.ResponseWriter, value any) { writeJSON(w, http.StatusCreated, value) }
func writeError(w http.ResponseWriter, err error) {
	status :=
		http.StatusInternalServerError
	switch {
	case errors.Is(err, domain.ErrInvalid):
		status =
			http.StatusBadRequest
	case errors.Is(err, domain.ErrUnauthorized):
		status =
			http.StatusUnauthorized
	case errors.Is(err, domain.ErrMissing):
		status =
			http.StatusNotFound
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrCapacity):
		status =
			http.StatusConflict
	}
	http.Error(w, `{"error":"`+err.Error()+`"}`, status)
}
