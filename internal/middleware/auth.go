package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Kostaaa1/tinylink/internal/common/authcontext"
	"github.com/Kostaaa1/tinylink/internal/domain/auth"
	"github.com/Kostaaa1/tinylink/pkg/errorhandler"
)

type MW struct {
	errorhandler.ErrorHandler
	// authService auth.Service
}

// PREP
// protected route middleware
// get jwt token from request
// verify token
// if its expired, get refresh token from cookie, validate it by checking if its expired
// if refresh token is expired force logout
// if not - then validate it by checking if it matches the user id from expired jwt token
// if refresh token is valid, generate new jwt, set header and when writing json extract it from header and include it json response?
// if refresh token is not valid send 403 Forbidden
func (mw MW) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearer := r.Header.Get("Authorization")
		accessToken := strings.TrimPrefix(bearer, "Bearer ")

		// cookie, err := r.Cookie(auth.SessionKey)
		// if err != nil {
		// 	// handle no refresh token
		// }
		// refreshToken := cookie.Value
		// claims, err := mw.authService.Authenticate(r.Context(), accessToken, refreshToken)

		claims, err := auth.VerifyAccessToken(accessToken)
		if err != nil {
			switch {
			case err == auth.ErrAccessTokenExpired:
				cookie, err := r.Cookie(auth.SessionKey)
				if err == http.ErrNoCookie {
					mw.UnauthorizedResponse(w, r)
					return
				}
				cookieStr, _ := json.MarshalIndent(cookie, "", " ")
				fmt.Println("Cookie: ", string(cookieStr))
				// if not - 403 Forbidden Error
				// refreshToken := cookie.Value
				// fmt.Println(refreshToken)
			default:
				mw.UnauthorizedResponse(w, r)
			}
			return
		}

		r = r.WithContext(authcontext.WithClaims(r.Context(), claims))
		next.ServeHTTP(w, r)
	})
}
