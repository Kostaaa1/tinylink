package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Kostaaa1/tinylink/internal/common/authcontext"
	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/Kostaaa1/tinylink/internal/domain/token"
	"github.com/Kostaaa1/tinylink/pkg/errorhandler"
)

type Auth struct {
	errorhandler.ErrorHandler
	tokenService *token.Service
}

func New(errHandler errorhandler.ErrorHandler, tokenService *token.Service) Auth {
	return Auth{
		ErrorHandler: errHandler,
		tokenService: tokenService,
	}
}

func (mw Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearer := r.Header.Get("Authorization")
		if bearer == "" {
			mw.UnauthorizedResponse(w, r)
			return
		}

		accessToken := strings.TrimPrefix(bearer, "Bearer ")

		claims, err := token.VerifyAccessToken(accessToken)
		if err == nil {
			r = r.WithContext(authcontext.WithClaims(r.Context(), claims))
			next.ServeHTTP(w, r)
			return
		}

		if errors.Is(err, token.ErrAccessTokenExpired) {
			cookie, err := r.Cookie(token.SessionKey)
			if err != nil {
				mw.UnauthorizedResponse(w, r)
				return
			}
			refreshToken := cookie.Value

			newRT, newAT, claims, err := mw.tokenService.RefreshTokens(r.Context(), refreshToken, claims.UserID)
			if err != nil {
				switch {
				case errors.Is(err, token.ErrTokenNotValid):
					mw.ForbiddenResponse(w, r)
				case errors.Is(err, data.ErrNotFound):
					http.SetCookie(w, &http.Cookie{
						Name:   token.SessionKey,
						Value:  "",
						MaxAge: -1,
						Path:   "/",
					})
					mw.UnauthorizedResponse(w, r)
				default:
					mw.UnauthorizedResponse(w, r)
				}
				return
			}

			w.Header().Set("Authorization", "Bearer "+newAT)
			r.AddCookie(&http.Cookie{
				Name:     token.SessionKey,
				Value:    newRT,
				Secure:   false, // true for https
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
				Domain:   "localhost", // change when in prod
				MaxAge:   int(token.RefreshTokenDuration.Seconds()),
			})

			r = r.WithContext(authcontext.WithClaims(r.Context(), claims))
			next.ServeHTTP(w, r)
			return
		}

		mw.UnauthorizedResponse(w, r)
	})
}
