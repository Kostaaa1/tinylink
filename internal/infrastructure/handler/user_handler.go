package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/Kostaaa1/tinylink/internal/domain/token"
	"github.com/Kostaaa1/tinylink/internal/domain/user"
	"github.com/Kostaaa1/tinylink/internal/middleware"
	"github.com/Kostaaa1/tinylink/pkg/errorhandler"
	"github.com/Kostaaa1/tinylink/pkg/validator"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type UserHandler struct {
	errorhandler.ErrorHandler
	userService  *user.Service
	oauth2Config *oauth2.Config
}

func NewUserHandler(userService *user.Service, errHandler errorhandler.ErrorHandler) UserHandler {
	return UserHandler{
		ErrorHandler: errHandler,
		userService:  userService,
		oauth2Config: &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_CALLBACK_URL"),
			Scopes:       []string{"profile", "email"},
			Endpoint:     google.Endpoint,
		},
	}
}

func (h UserHandler) RegisterRoutes(r *mux.Router, auth middleware.Auth) {
	r.HandleFunc("/login/google", h.HandleGoogleRedirect).Methods("GET")
	r.HandleFunc("/google/callback", h.HandleGoogleCallback).Methods("GET")
	//////////////////////////////////////////////////////////////////////////////////////////
	userRoutes := r.PathPrefix("/user").Subrouter()
	userRoutes.HandleFunc("/register", h.Register).Methods("POST")
	userRoutes.HandleFunc("/login", h.Login).Methods("POST")
	userRoutes.HandleFunc("/logout", h.Logout).Methods("POST")

	protected := r.PathPrefix("/user").Subrouter()
	protected.Use(auth.Middleware)
	// protected.HandleFunc("/refresh-token", h.HandleRefreshToken).Methods("POST")
	protected.HandleFunc("/change-password", h.ChangePassword).Methods("POST")
	// protected.HandleFunc("/refresh-token", h.HandleRefreshToken).Methods("GET")
}

func (h UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	cookie, _ := r.Cookie(token.SessionKey)
	if err := h.userService.Logout(ctx, cookie.Value); err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, envelope{"message": "successful logout"}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h UserHandler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
}

func (h UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Password string `json:"password"`
	}
	if err := readJSON(r, &input); err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	user.ValidatePasswordPlainText(v, input.Password)
	if !v.Valid() {
		h.FailedValidationResponse(w, r, v.Errors)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.userService.ChangePassword(ctx, input.Password); err != nil {
		if errors.Is(err, data.ErrNotFound) {
			h.NotFoundResponse(w, r)
			return
		}
		h.ServerErrorResponse(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, nil, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}
}

func (h UserHandler) HandleGoogleRedirect(w http.ResponseWriter, r *http.Request) {
	// TODO:
	url := h.oauth2Config.AuthCodeURL("random-state", oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h UserHandler) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	code := r.URL.Query().Get("code")

	googleToken, err := h.oauth2Config.Exchange(ctx, code)
	if err != nil {
		h.ServerErrorResponse(w, r, fmt.Errorf("failed to exchange token: %v", err))
		return
	}

	client := h.oauth2Config.Client(ctx, googleToken)

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

	loggedUser, err := h.userService.HandleGoogleLogin(ctx, &googleUser)
	if err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	token, _, err := token.GenerateAccessToken(loggedUser.ID)
	if err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	if err := writeJSON(w, http.StatusOK, envelope{"user": loggedUser, "token": token}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input user.LoginRequest
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

	userData, err := h.userService.Login(ctx, input.Email, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidCredentials):
			h.InvalidCredentialsResponse(w, r)
		case errors.Is(err, data.ErrNotFound):
			h.NotFoundResponse(w, r)
		case errors.Is(err, user.ErrNoUserPasswordSet):
			h.BadRequestResponse(w, r, user.ErrNoUserPasswordSet)
		default:
			h.ServerErrorResponse(w, r, err)
		}
		return
	}

	token, _, err := token.GenerateAccessToken(userData.ID)
	if err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	if err := writeJSON(w, http.StatusOK, envelope{"user": userData, "token": token}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req user.RegisterRequest
	if err := readJSON(r, &req); err != nil {
		h.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	v := validator.New()
	user.ValidateEmail(v, req.Email)
	user.ValidatePasswordPlainText(v, req.Password)
	if !v.Valid() {
		h.FailedValidationResponse(w, r, v.Errors)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	dto, err := h.userService.Register(ctx, &req)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrDuplicateEmail):
			h.ErrorResponse(w, r, http.StatusConflict, "user already exists with this email address")
		default:
			h.ServerErrorResponse(w, r, err)
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, envelope{"user": dto}, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}
