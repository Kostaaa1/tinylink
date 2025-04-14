package token

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	sessionKey = "tinylink_session"
	jwtSecret  = []byte(os.Getenv("JWT_SECRET_KEY"))

	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 7 * 24 * time.Hour

	ErrAccessTokenExpired  = errors.New("access token expired")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrTokenNotValid       = errors.New("token user id does not match provided user id")
)

type Claims struct {
	UserID string
	JTI    string // blacklisting
	jwt.RegisteredClaims
}

// used in unprotected routes where user claims could be nil (tinylink.Create())
func ClaimsFromRequest(r *http.Request) (*Claims, error) {
	bearer := r.Header.Get("Authorization")
	token := strings.TrimPrefix(bearer, "Bearer ")
	return VerifyAccessToken(token)
}

func GenerateRefreshToken() string {
	return uuid.NewString()
}

func SetHeaderAndCookie(w http.ResponseWriter, r *http.Request, refreshToken, accessToken string) {
	w.Header().Set("Authorization", "Bearer "+accessToken)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionKey,
		Value:    refreshToken,
		Secure:   false, // true for https
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   int(refreshTokenDuration.Seconds()),
	})
}

func ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   sessionKey,
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})
}

func GetRefreshToken(r *http.Request) (string, error) {
	token, err := r.Cookie(sessionKey)
	if err != nil {
		return "", err
	}
	return token.Value, nil
}

func GenerateAccessToken(userID uint64) (string, *Claims, error) {
	id := strconv.FormatUint(userID, 10)

	claims := &Claims{
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
		return "", nil, err
	}

	return signed, claims, nil
}

func VerifyAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrAccessTokenExpired
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, errors.New("access token not found??")
}
