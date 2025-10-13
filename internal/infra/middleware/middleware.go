package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Kostaaa1/tinylink/internal/constants"
	"github.com/Kostaaa1/tinylink/internal/domain/auth"
	"github.com/Kostaaa1/tinylink/internal/domain/token"
	"github.com/Kostaaa1/tinylink/pkg/errhandler"
)

type Middleware struct {
	errHandler errhandler.ErrorHandler
	token      *token.Service
	log        *slog.Logger
}

func New(errHandler errhandler.ErrorHandler, token *token.Service, log *slog.Logger) Middleware {
	return Middleware{
		errHandler: errHandler,
		token:      token,
		log:        log,
	}
}

func (mw Middleware) RouteProtector(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// userCtx, err := userCtxFromClaims(w, r)
		// if err == nil {
		// 	r = r.WithContext(auth.WithClaims(r.Context(), userCtx))
		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		// needs to use context only, validation of JWT access token needs to happen in GlobalMiddleware.
		claims, err := auth.ClaimsFromRequest(r)
		if err == nil {
			next.ServeHTTP(w, r)
			return
		}

		switch {
		case errors.Is(err, auth.ErrTokenNotProvided):
			mw.errHandler.UnauthorizedResponse(w, r)
			return
		case errors.Is(err, auth.ErrAccessTokenExpired):
			oldToken, err := auth.RefreshTokenFromCookie(r)
			if err != nil {
				mw.errHandler.ErrorResponse(w, r, http.StatusUnauthorized, err)
				return
			}

			refresh, access, err := mw.token.ValidateAndRotateTokens(r.Context(), claims.UserID, oldToken)
			if err != nil {
				switch {
				case errors.Is(err, auth.ErrTokenNotValid):
					mw.errHandler.ForbiddenResponse(w, r)
				case errors.Is(err, constants.ErrNotFound):
					auth.ClearRefreshToken(w)
					mw.errHandler.UnauthorizedResponse(w, r)
				default:
					mw.errHandler.UnauthorizedResponse(w, r)
				}
				return
			}

			auth.SetTokens(w, refresh, access)
			r = r.WithContext(auth.WithClaims(r.Context(), auth.UserContext{
				IsAuthenticated: true,
				UserID:          &claims.UserID,
				GuestUUID:       auth.EnsureGuestUUID(w, r),
				Roles:           claims.Roles,
			}))

			next.ServeHTTP(w, r)
			return
		}

		mw.errHandler.UnauthorizedResponse(w, r)
	})
}

func (mw Middleware) verifyCSRFToken(r *http.Request, guestUUID string) bool {
	sess := mw.token.GetSession(r.Context(), guestUUID)
	token := r.FormValue("csrf_token")
	if token == "" {
		token = r.Header.Get("X-XSRF-Token")
	}
	return sess.CSRF == token
}

// What needs to happen:
// 1. get guest uuid from cookie, create and store if does not exist..
// 2. get session data from redis (CSRF token)
// 3. If method is POST, PUT, PATCH, DELETE - check if CSRF from redis matches the one from the request.
// 4. check if there is user JWT Bearer token, if it exists validate it.
// 5. create UserContext and store it in r = r.Context
// 6. proceed
func (mw Middleware) Global(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		guestUUID := auth.EnsureGuestUUID(w, r)

		// ensures that caches, browser caches or CDNs, differentiates responses based on the presence or value of the Cookie Header.
		w.Header().Add("Vary", "Cookie")
		// this tells caches not to store responses that include the Set-Cookie header
		w.Header().Add("Cache-Control", `no-cache="Set-Cookie"`)

		// Validate CSRF token
		if r.Method == http.MethodPost ||
			r.Method == http.MethodPut ||
			r.Method == http.MethodPatch ||
			r.Method == http.MethodDelete {
			if !mw.verifyCSRFToken(r, guestUUID) {
				mw.errHandler.ErrorResponse(w, r, http.StatusForbidden, "CSRF token mismatch")
				return
			}
		}

		claims, err := auth.ClaimsFromRequest(r)
		userCtx := auth.UserContext{
			GuestUUID:       guestUUID,
			IsAuthenticated: err == nil,
			Error:           err,
			Roles:           claims.Roles,
			UserID:          &claims.UserID,
		}

		r = r.WithContext(auth.WithClaims(r.Context(), userCtx))
		next.ServeHTTP(w, r)
	})
}
