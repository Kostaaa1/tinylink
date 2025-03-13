package handler

import (
	"net/http"
)

type Handler struct {
	*ErrorHandler
	User     *UserHandler
	Tinylink *TinylinkHandler
}

func (h *Handler) HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": "devlopment",
			"version":     "1.0",
		},
	}
	err := writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}
