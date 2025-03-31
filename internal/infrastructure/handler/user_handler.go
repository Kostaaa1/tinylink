package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/Kostaaa1/tinylink/internal/common/validator"
	"github.com/Kostaaa1/tinylink/internal/domain/user"
	"github.com/coreos/go-oidc"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type UserHandler struct {
	*ErrorHandler
	service *user.Service
}

func NewUserHandler(userService *user.Service, errHandler *ErrorHandler) *UserHandler {
	return &UserHandler{
		ErrorHandler: errHandler,
		service:      userService,
	}
}

func (h *UserHandler) RegisterRoutes(r *mux.Router) {
	oauth2Config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_CALLBACK_URL"),
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}

	oidcProvider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		log.Fatalf("Failed to get provider: %v", err)
	}

	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		url := oauth2Config.AuthCodeURL("random-state", oauth2.AccessTypeOffline)
		http.Redirect(w, r, url, http.StatusFound)
	})

	r.HandleFunc(os.Getenv("GOOGLE_CALLBACK_URL"), func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		code := r.URL.Query().Get("code")

		token, err := oauth2Config.Exchange(ctx, code)
		if err != nil {
			http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
			return
		}

		// Verify ID token
		verifier := oidcProvider.Verifier(&oidc.Config{ClientID: oauth2Config.ClientID})
		idToken, err := verifier.Verify(ctx, token.Extra("id_token").(string))
		if err != nil {
			http.Error(w, "Failed to verify ID token", http.StatusInternalServerError)
			return
		}

		// Get user info
		var userInfo struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		}
		if err := idToken.Claims(&userInfo); err != nil {
			http.Error(w, "Failed to parse claims", http.StatusInternalServerError)
			return
		}

		// Return user info
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userInfo)
	})

	////////////////////////////////////////////////////////////////////////////////////////////////

	// userRoutes := r.PathPrefix("/users").Subrouter()
	// userRoutes.HandleFunc("/register", h.Register).Methods("POST")
	// userRoutes.HandleFunc("/login", h.Login).Methods("POST")
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := readJSON(r, &input); err != nil {
		h.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	user.ValidateEmail(v, input.Email)
	user.ValidatePasswordPlainText(v, input.Password)
	if !v.Valid() {
		h.FailedValidationResponse(w, r, v.Errors)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	userData, err := h.service.Login(ctx, input.Email, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidCredentials):
			h.InvalidCredentialsResponse(w, r)
		case errors.Is(err, data.ErrRecordNotFound):
			h.NotFoundResponse(w, r)
		default:
			h.ServerErrorResponse(w, r, err)
		}
		return
	}

	// get token or generate and return it...

	if err := writeJSON(w, http.StatusOK, envelope{"user": userData}, nil); err != nil {
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

	userData := &user.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err := userData.Password.Set(input.Password)
	if err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	if user.ValidateUser(v, userData); !v.Valid() {
		h.FailedValidationResponse(w, r, v.Errors)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	err = h.service.Register(ctx, userData)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrDuplicateEmail):
			v.AddError("email", "user already exists with this email address.")
			h.FailedValidationResponse(w, r, v.Errors)
		default:
			h.ServerErrorResponse(w, r, err)
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, envelope{"user": userData}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}
