package token

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	SessionTTL             = 365 * 24 * time.Hour
	sessionKey             = "tinylink_session"
	refreshTokenKey        = "refresh_token"
	jwtSecret              = []byte(os.Getenv("JWT_SECRET_KEY"))
	accessTokenDuration    = 15 * time.Minute
	refreshTokenDuration   = 7 * 24 * time.Hour
	ErrAccessTokenExpired  = errors.New("access token expired")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrTokenNotValid       = errors.New("token user id does not match provided user id")
)

type Claims struct {
	UserID string
	JTI    string // uuid for blacklisting
	jwt.RegisteredClaims
}

// used in unprotected routes where user claims could be nil (tinylink.Create())
func ClaimsFromRequest(r *http.Request) (Claims, error) {
	bearer := r.Header.Get("Authorization")
	token := strings.TrimPrefix(bearer, "Bearer ")
	return VerifyAccessToken(token)
}

func GenerateRefreshToken() string {
	return uuid.NewString()
}

func SetHeaderAndCookie(w http.ResponseWriter, refreshToken, accessToken string) {
	w.Header().Set("Authorization", "Bearer "+accessToken)
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenKey,
		Value:    refreshToken,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   int(refreshTokenDuration.Seconds()),
	})
}

func ClearRefreshToken(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   refreshTokenKey,
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})
}

func GetSessionID(r *http.Request) (string, error) {
	token, err := r.Cookie(sessionKey)
	if err != nil {
		return "", err
	}
	return token.Value, nil
}

// Retrieves or creates a new session ID with a 1-year expiration
func GetOrCreateSessionID(w http.ResponseWriter, r *http.Request) (string, error) {
	token, err := r.Cookie(sessionKey)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			sessID := uuid.NewString()
			http.SetCookie(w, &http.Cookie{
				Name:     sessionKey,
				Value:    sessID,
				Secure:   false,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
				MaxAge:   int(SessionTTL.Seconds()),
			})
			return sessID, nil
		}

		return "", fmt.Errorf("failed to read session cookie: %w", err)
	}
	return token.Value, nil
}

func GetRefreshToken(r *http.Request) (string, error) {
	token, err := r.Cookie(refreshTokenKey)
	if err != nil {
		return "", err
	}
	return token.Value, nil
}

// this sucks
func GetAuthIdentifiers(w http.ResponseWriter, r *http.Request) (string, string, error) {
	claims, err := ClaimsFromRequest(r)
	if err != nil {
		return "", "", err
	}
	sessionID, err := GetOrCreateSessionID(w, r)
	if err != nil {
		return "", "", err
	}
	if sessionID == "" && claims.UserID == "" {
		return "", "", data.ErrUnauthenticated
	}
	return claims.UserID, sessionID, nil
}

func GenerateAccessToken(userID uint64) (string, Claims, error) {
	id := strconv.FormatUint(userID, 10)

	claims := Claims{
		UserID: id,
		JTI:    uuid.NewString(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", Claims{}, err
	}

	return signed, claims, nil
}

func VerifyAccessToken(tokenStr string) (Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return Claims{}, err
	}
	if !token.Valid {
		return Claims{}, ErrAccessTokenExpired
	}
	if claims, ok := token.Claims.(*Claims); ok {
		return *claims, nil
	}
	return Claims{}, errors.New("access token not found??")
}
