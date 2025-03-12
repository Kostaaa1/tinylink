package handlers

import (
	"net/http"
)

func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": "devlopment",
			"version":     "1.0",
		},
	}
	err := writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		ServerErrorResponse(w, r, err)
	}
}
