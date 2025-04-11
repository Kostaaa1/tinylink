package handler

import (
	"net/http"

	"github.com/Kostaaa1/tinylink/pkg/errorhandler"
)

type Handler struct {
	errorhandler.ErrorHandler
	User     UserHandler
	Tinylink TinylinkHandler
}

func (h Handler) HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
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
