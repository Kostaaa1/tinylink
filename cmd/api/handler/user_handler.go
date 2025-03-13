package handler

import (
	"errors"
	"net/http"

	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/Kostaaa1/tinylink/internal/services"
	"github.com/Kostaaa1/tinylink/internal/validator"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	*ErrorHandler
	service *services.UserService
}

func NewUserHandler(userService *services.UserService, errHandler *ErrorHandler) *UserHandler {
	return &UserHandler{
		ErrorHandler: errHandler,
		service:      userService,
	}
}

func (h *UserHandler) RegisterRoutes(r *mux.Router) {
	userRoutes := r.PathPrefix("/users").Subrouter()
	userRoutes.HandleFunc("/register", h.Register).Methods("POST")
	userRoutes.HandleFunc("/{email}", h.GetByEmail).Methods("GET")
}

func (h *UserHandler) GetByEmail(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var email string
	if len(params) > 0 {
		email = params["email"]
	}

	user, err := h.service.GetByEmail(email)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			h.NotFoundResponse(w, r)
			return
		}
		h.ServerErrorResponse(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, envelope{"user": user}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := readJSON(r, &input); err != nil {
		h.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err := user.Password.Set(input.Password)
	if err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		h.FailedValidationResponse(w, r, v.Errors)
		return
	}

	// err = h.Register(user)
	err = h.service.Register(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "user already exists with this email address.")
			h.FailedValidationResponse(w, r, v.Errors)
		default:
			h.ServerErrorResponse(w, r, err)
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, envelope{"user": user}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}
