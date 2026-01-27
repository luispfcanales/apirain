package response

import (
	"encoding/json"
	"net/http"
)

func SetupCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	SetupCORS(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func InternalServerError(w http.ResponseWriter, message string) {
	SetupCORS(w)
	JSON(w, http.StatusInternalServerError, map[string]string{"error": message})
}
