package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kostaaa1/tinylink/internal/common/auth"
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

func (s *Service) HandleGoogleLogin(ctx context.Context, googleUser *GoogleUser) (UserDTO, error) {
	return s.user.HandleGoogleLogin(ctx, googleUser)
}

func (s *Service) GetUserFromCtx(ctx context.Context) (*User, error) {
	claims := auth.ClaimsFromCtx(ctx)
	return s.user.GetByEmail(ctx, claims.Email)
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

func (s *Service) Register(ctx context.Context, user *User) error {
	newUser, err := s.user.GetByEmail(ctx, user.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return s.user.Insert(ctx, user)
		default:
			return err
		}
	}
	if newUser != nil {
		fmt.Println("updating user")
		return s.user.Update(ctx, user)
	}
	return nil
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
