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
	status := responseStatusForError(err)
	http.Error(w, `{"error":"`+err.Error()+`"}`, status)
}

func responseStatusForError(err error) int {
	if errors.Is(err, domain.ErrInvalid) {
		return http.StatusBadRequest
	}
	if errors.Is(err, domain.ErrUnauthorized) {
		return http.StatusUnauthorized
	}
	if errors.Is(err, domain.ErrMissing) {
		return http.StatusNotFound
	}
	if errors.Is(err, domain.ErrConflict) || errors.Is(err, domain.ErrCapacity) {
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}
