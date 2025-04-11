package auth

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
	SessionKey             = "tinylink_session"
	jwtSecret              = []byte(os.Getenv("JWT_SECRET_KEY"))
	AccessTokenDuration    = 15 * time.Minute
	RefreshTokenDuration   = 7 * 24 * time.Hour
	ErrAccessTokenExpired  = errors.New("access token expired")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrTokenNotValid       = errors.New("token does not match user id")
)

type RefreshToken struct {
	ID        string
	UserID    string
	CreatedAt time.Duration
	ExpiresAt time.Duration
}

type Claims struct {
	UserID string
	Email  string
	jwt.RegisteredClaims
}

// used in unprotected routes that could need user claims (tinylink.Create())
func GetClaimsFromRequest(r *http.Request) (*Claims, error) {
	bearer := r.Header.Get("Authorization")
	token := strings.TrimPrefix(bearer, "Bearer ")
	if token == "" {
		cookie, _ := r.Cookie(SessionKey)
		token = cookie.Value
	}
	return VerifyAccessToken(token)
}

func GenerateRefreshToken() string {
	return uuid.NewString()
}

func GenerateAccessToken(userID uint64, email string, dur time.Duration) (string, error) {
	id := strconv.FormatUint(userID, 10)
	claims := &Claims{
		UserID: id,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(dur)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
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
