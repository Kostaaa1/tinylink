package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Kostaaa1/tinylink/internal/common/auth"
	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/Kostaaa1/tinylink/internal/common/validator"
	"github.com/Kostaaa1/tinylink/internal/domain/user"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type UserHandler struct {
	*ErrorHandler
	service      *user.Service
	oauth2Config *oauth2.Config
}

func NewUserHandler(userService *user.Service, errHandler *ErrorHandler) *UserHandler {
	return &UserHandler{
		ErrorHandler: errHandler,
		service:      userService,
		oauth2Config: &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_CALLBACK_URL"),
			Scopes:       []string{"profile", "email"},
			Endpoint:     google.Endpoint,
		},
	}
}

func (h *UserHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/login/google", h.HandleGoogleRedirect).Methods("GET")
	r.HandleFunc("/auth/google/callback", h.HandleGoogleCallback).Methods("GET")
	//////////////////////////////////////////////////////////////////////////////////////////
	userRoutes := r.PathPrefix("/user").Subrouter()
	userRoutes.HandleFunc("/register", h.Register).Methods("POST")
	userRoutes.HandleFunc("/login", h.Login).Methods("POST")
}

func (h *UserHandler) HandleGoogleRedirect(w http.ResponseWriter, r *http.Request) {
	url := h.oauth2Config.AuthCodeURL("random-state", oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *UserHandler) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	code := r.URL.Query().Get("code")

	// handle csrf attacks

	token, err := h.oauth2Config.Exchange(ctx, code)
	if err != nil {
		h.ServerErrorResponse(w, r, fmt.Errorf("failed to exchange token: %v", err))
		return
	}

	client := h.oauth2Config.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		h.BadRequestResponse(w, r, fmt.Errorf("failed to exchange token: %v", err))
		return
	}
	defer resp.Body.Close()

	var googleUser user.GoogleUser
	err = json.NewDecoder(resp.Body).Decode(&googleUser)
	if err != nil {
		h.ServerErrorResponse(w, r, fmt.Errorf("failed to exchange token: %v", err))
		return
	}

	loggedUser, err := h.service.HandleGoogleLogin(ctx, &googleUser)
	if err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	// TODO: should create if not exists user from google user data
	// google - upon login - check if users table have the provided email - if it does, then automatically link 2 accounts.
	// normal - registration - prompt them for credentials, if google_users has that email link it automatically
	// normal - login - if not found, check if there is an email in google_users, if exists prompt them to add password and create row in users table
	// linking process - if not exists, insert row in users/google_users table and save user_id in google_users

	jwtToken, err := auth.GenerateJWT(loggedUser.ID, loggedUser.Email)
	if err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, envelope{"data": loggedUser, "token": jwtToken}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
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

	token, err := auth.GenerateJWT(userData.ID, userData.Email)
	if err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, envelope{"user": userData, "token": token}, nil); err != nil {
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
		Name:  input.Name,
		Email: input.Email,
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
			h.ErrorResponse(w, r, http.StatusConflict, "user already exists with this email address")
		default:
			h.ServerErrorResponse(w, r, err)
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, envelope{"user": userData}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}
