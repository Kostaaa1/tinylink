package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Kostaaa1/tinylink/internal/constants"
	"github.com/Kostaaa1/tinylink/internal/domain/auth"
	"github.com/Kostaaa1/tinylink/internal/domain/user"
	"github.com/Kostaaa1/tinylink/pkg/errhandler"
	"github.com/Kostaaa1/tinylink/pkg/jsonutil"
	"github.com/Kostaaa1/tinylink/pkg/validator"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	oauth2State = "random-state-idk"
)

type UserHandler struct {
	errhandler.ErrorHandler
	userService  *user.Service
	oauth2Config *oauth2.Config
	log          *slog.Logger
}

func NewUserHandler(userService *user.Service, errHandler errhandler.ErrorHandler, log *slog.Logger) UserHandler {
	return UserHandler{
		ErrorHandler: errHandler,
		userService:  userService,
		log:          log,
		oauth2Config: &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_CALLBACK_URL"),
			Scopes:       []string{"profile", "email"},
			Endpoint:     google.Endpoint,
		},
	}
}

func (h UserHandler) RegisterRoutes(r *mux.Router, requireAuthMW mux.MiddlewareFunc) {
	r.HandleFunc("/login/google", h.HandleGoogleRedirect).Methods("GET")
	r.HandleFunc("/auth/google/callback", h.HandleGoogleCallback).Methods("GET")

	userRoutes := r.PathPrefix("/user").Subrouter()
	userRoutes.HandleFunc("/register", h.Register).Methods("POST")
	userRoutes.HandleFunc("/login", h.Login).Methods("POST")

	protected := r.PathPrefix("/user").Subrouter()
	protected.Use(requireAuthMW)
	protected.HandleFunc("/change-password", h.ChangePassword).Methods("PATCH")
	protected.HandleFunc("/logout", h.Logout).Methods("POST")

	// TESTING ONLY - REMOVE LATER
	// protected.HandleFunc("/refresh-token", h.HandleRefreshToken).Methods("GET")
}

func (h UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.FromContext(r.Context())
	if !userCtx.IsAuthenticated {
		h.UnauthorizedResponse(w, r)
		return
	}

	if err := h.userService.Logout(r.Context(), *userCtx.UserID); err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	auth.ClearRefreshToken(w)

	if err := jsonutil.Write(w, http.StatusOK, "successful logout", nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var input struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := jsonutil.Read(r, &input); err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	if _ = user.ValidatePasswordPlainTextWithKey(v, "new_password", input.NewPassword); !v.Valid() {
		h.FailedValidationResponse(w, r, v.Errors)
		return
	}

	userCtx := auth.FromContext(r.Context())
	if !userCtx.IsAuthenticated {
		h.UnauthorizedResponse(w, r)
		return
	}

	if err := h.userService.ChangePassword(r.Context(), *userCtx.UserID, input.OldPassword, input.NewPassword); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			h.NotFoundResponse(w, r)
			return
		}
		h.ServerErrorResponse(w, r, err)
		return
	}

	if err := jsonutil.Write(w, http.StatusOK, "password changed successfully", nil); err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}
}

func (h UserHandler) HandleGoogleRedirect(w http.ResponseWriter, r *http.Request) {
	// TODO:
	url := h.oauth2Config.AuthCodeURL(oauth2State, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h UserHandler) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	googleToken, err := h.oauth2Config.Exchange(r.Context(), code)
	if err != nil {
		h.ServerErrorResponse(w, r, fmt.Errorf("failed to exchange token: %v", err))
		return
	}

	client := h.oauth2Config.Client(r.Context(), googleToken)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		h.BadRequestResponse(w, r, fmt.Errorf("failed to fetch user google data: %v", err))
		return
	}
	defer resp.Body.Close()

	var googleUser user.GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		h.ServerErrorResponse(w, r, fmt.Errorf("failed to read JSON: %v", err))
		return
	}

	loggedUser, err := h.userService.HandleGoogleLogin(r.Context(), &googleUser)
	if err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	token, err := auth.GenerateAccessToken(loggedUser.ID, nil)
	if err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)

	response := jsonutil.Envelope{
		"user":  UserResponse(loggedUser),
		"token": token,
	}

	if err := jsonutil.Write(w, http.StatusOK, response, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input UserLoginRequest
	if err := jsonutil.Read(r, &input); err != nil {
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

	loggedUser, accessToken, refreshToken, err := h.userService.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidCredentials):
			h.InvalidCredentialsResponse(w, r)
		case errors.Is(err, constants.ErrNotFound):
			h.NotFoundResponse(w, r)
		case errors.Is(err, user.ErrNoUserPasswordSet):
			h.BadRequestResponse(w, r, user.ErrNoUserPasswordSet)
		default:
			h.ServerErrorResponse(w, r, err)
		}
		return
	}

	auth.SetTokens(w, refreshToken, accessToken)

	resp := jsonutil.Envelope{
		"user":  UserResponse(loggedUser),
		"token": accessToken,
	}

	if err := jsonutil.Write(w, http.StatusOK, resp, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}

func (h UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req UserRegisterRequest
	if err := jsonutil.Read(r, &req); err != nil {
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

	userData := &user.User{
		Name:  req.Name,
		Email: req.Email,
	}

	if err := userData.Password.Set(req.Password); err != nil {
		h.ServerErrorResponse(w, r, err)
		return
	}

	err := h.userService.Register(r.Context(), userData)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrDuplicateEmail):
			h.ErrorResponse(w, r, http.StatusConflict, "user already exists with this email address")
		default:
			h.ServerErrorResponse(w, r, err)
		}
		return
	}

	resp := jsonutil.Envelope{"user": UserResponse(userData)}

	if err := jsonutil.Write(w, http.StatusOK, resp, nil); err != nil {
		h.ServerErrorResponse(w, r, err)
	}
}
