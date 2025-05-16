package utils

import (
	"encoding/json"
	"net/http"
)

type APIResponse[T any] struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    *T     `json:"data,omitempty"`
}

func RespondWithJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}
