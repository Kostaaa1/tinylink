package services

import (
	"context"
	"strconv"
	"time"

	"github.com/Kostaaa1/tinylink/db"
	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/Kostaaa1/tinylink/internal/middleware/auth"
)

type UserService struct {
	User  db.UserStore
	Token db.TokenStore
}

func NewUserService(userStore db.UserStore, tokenStore db.TokenStore) *UserService {
	return &UserService{
		User:  userStore,
		Token: tokenStore,
	}
}

func (s *UserService) Login(ctx context.Context, email, password string) (*data.User, *data.Token, error) {
	user, err := s.User.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}

	matches, err := user.Password.Matches(password)
	if err != nil {
		return nil, nil, err
	}

	if !matches {
		return nil, nil, data.ErrInvalidCredentials
	}

	token := auth.AuthTokenFromContext(ctx)
	if token != nil {
		sessionTTL := time.Hour * 24 * 30
		userID := strconv.FormatUint(user.ID, 10)
		token, err = data.GenerateToken(userID, data.DefaultTokenTTL, data.ScopeAuthentication)
		if err != nil {
			return nil, nil, err
		}
		if err := s.Token.Store(ctx, token, sessionTTL); err != nil {
			return nil, nil, err
		}
	}

	return user, token, nil
}

func (s *UserService) Register(ctx context.Context, user *data.User) error {
	return s.User.Insert(ctx, user)
}
