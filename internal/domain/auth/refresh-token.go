package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
)

func SetTokens(w http.ResponseWriter, refreshToken, accessToken string) {
	w.Header().Set("Authorization", "Bearer "+accessToken)
	setRefreshToken(w)
}

func ClearRefreshToken(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   refreshTokenKey,
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})
}

func setRefreshToken(w http.ResponseWriter) string {
	token := GenerateRefreshToken()
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenKey,
		Value:    token,
		Secure:   true,                    // if true, ensures that the cookie is only sent over HTTPS
		HttpOnly: true,                    // prevents javascript from accessing the cookie
		SameSite: http.SameSiteStrictMode, // restricts cross-site cookie transmission
		Path:     "/",
		MaxAge:   int(RefreshTokenTTL.Seconds()),
	})
	return token
}

func EnsureRefreshToken(w http.ResponseWriter, r *http.Request) (string, error) {
	token, err := r.Cookie(refreshTokenKey)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return setRefreshToken(w), nil
		}
		return "", fmt.Errorf("failed to read session cookie: %w", err)
	}
	return token.Value, nil
}

func RefreshTokenFromCookie(r *http.Request) (string, error) {
	token, err := r.Cookie(refreshTokenKey)
	if err != nil {
		return "", err
	}
	if token.Value == "" {
		return "", errors.New("session cookie value missing")
	}
	return token.Value, nil
}

func GenerateRefreshToken() string {
	buf := make([]byte, 16)
	rand.Read(buf)
	return base64.RawURLEncoding.EncodeToString(buf)
}
