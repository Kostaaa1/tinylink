package user

import (
	"context"

	"github.com/Kostaaa1/tinylink/internal/domain/token"
)

type Service struct {
	User  Repository
	Token token.Repository
}

func NewService(userRepo Repository, tokenRepo token.Repository) *Service {
	return &Service{
		User:  userRepo,
		Token: tokenRepo,
	}
}

func (s *Service) Login(ctx context.Context, email, password string) (*User, error) {
	userData, err := s.User.GetByEmail(ctx, email)
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
	return s.User.Insert(ctx, user)
}
