package auth

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

var (
	jwtSecret                 = []byte(os.Getenv("JWT_SECRET_KEY"))
	userContextKey contextKey = "user_context"

	ErrAccessTokenExpired  = errors.New("access token expired")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrTokenNotValid       = errors.New("token user id does not match provided user id")
	ErrTokenNotProvided    = errors.New("bearer token not provided")

	AccessTokenDuration = 15 * time.Minute
	refreshTokenKey     = "tl_refresh_token"
	RefreshTokenTTL     = time.Hour * 24 * 7
)

type UserContext struct {
	IsAuthenticated bool
	UserID          *uint64
	GuestUUID       string
	Roles           []string
	Error           error
}

type Claims struct {
	UserID uint64
	Roles  []string
	jwt.RegisteredClaims
}

// Need function for protected routes, that will get me user ID and its roles without nullable values
// Need a function that will give me nullable userID and roles for unprotected routes
func WithClaims(ctx context.Context, userCtx UserContext) context.Context {
	return context.WithValue(ctx, userContextKey, userCtx)
}

// used only in protected routes (middleware set context value)
func FromContext(ctx context.Context) UserContext {
	userCtx, ok := ctx.Value(userContextKey).(UserContext)
	if !ok {
		panic("failed to assert context.Value to Claims struct")
	}
	return userCtx
}

// Gets Bearer token from request and validates it
func ClaimsFromRequest(r *http.Request) (Claims, error) {
	bearer := r.Header.Get("Authorization")

	token := strings.TrimPrefix(bearer, "Bearer ")
	if token == "" {
		return Claims{}, ErrTokenNotProvided
	}

	return ParseToken(token)
}

// Parse cookie as claims struct. Used for validation of jwt token
func ParseToken(tokenStr string) (Claims, error) {
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

func GenerateAccessToken(userID uint64, roles []string) (string, error) {
	if userID < 1 {
		return "", errors.New("missing userID")
	}

	claims := Claims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return signed, nil
}
