package router

import (
	"encoding/json"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]any{
		"name":    "School Management API",
		"status":  "running",
		"version": "1.0",
		"docs":    "https://github.com/burhaanAshraf/school-management-api",
	})
}
