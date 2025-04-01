package user

import (
	"context"
	"errors"

	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/Kostaaa1/tinylink/internal/domain/token"
)

type Service struct {
	user  Repository
	token token.Repository
}

func NewService(userRepo Repository, tokenRepo token.Repository) *Service {
	return &Service{
		user:  userRepo,
		token: tokenRepo,
	}
}

func (s *Service) FindOrCreate(ctx context.Context, user *User) (*User, error) {
	newUser, err := s.user.GetByEmail(ctx, user.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			if err := s.user.Insert(ctx, user); err != nil {
				return nil, err
			}
			return user, nil
		default:
			return nil, err
		}
	}
	return newUser, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*User, error) {
	userData, err := s.user.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	matches, err := userData.Password.Matches(password)
	if err != nil {
		return nil, err
	}
	if !matches {
		return nil, ErrInvalidCredentials
	}
	return userData, err
}

func (s *Service) Register(ctx context.Context, user *User) error {
	return s.user.Insert(ctx, user)
}
