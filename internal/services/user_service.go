package services

import (
	"context"
	"strconv"

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

	userID := strconv.FormatUint(user.ID, 10)
	token := auth.AuthTokenFromCtx(ctx)
	if token == nil {
		err = s.Token.RevokeAll(ctx, userID, &data.ScopeAuthentication)
		if err != nil {
			return nil, nil, err
		}
		token = data.GenerateToken(userID)
		if err := s.Token.Store(ctx, token); err != nil {
			return nil, nil, err
		}
	}

	return user, token, err
}

func (s *UserService) Register(ctx context.Context, user *data.User) error {
	return s.User.Insert(ctx, user)
}
