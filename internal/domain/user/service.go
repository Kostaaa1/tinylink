package user

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/Kostaaa1/tinylink/internal/common/authcontext"
	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/Kostaaa1/tinylink/internal/domain/token"
)

type Adapters struct {
	UserDbRepository UserRepository
}

type provider interface {
	WithTransaction(txFunc func(adapters Adapters) error) error
	GetAdapters() Adapters
}

type Service struct {
	provider  provider
	userDb    UserRepository
	tokenRepo token.Repository
}

func NewService(provider provider, tokenRepo token.Repository) *Service {
	adapters := provider.GetAdapters()
	return &Service{
		provider:  provider,
		userDb:    adapters.UserDbRepository,
		tokenRepo: tokenRepo,
	}
}

func (s *Service) HandleGoogleLogin(ctx context.Context, googleUser *GoogleUser) (*User, error) {
	user := &User{
		Name:   googleUser.Name,
		Email:  googleUser.Email,
		Google: googleUser,
	}

	err := s.provider.WithTransaction(func(adapters Adapters) error {
		existingUser, err := adapters.UserDbRepository.GetByEmail(ctx, user.Email)

		if err != nil {
			if errors.Is(err, data.ErrNotFound) {
				if err := adapters.UserDbRepository.Insert(ctx, user); err != nil {
					return fmt.Errorf("failed to insert user: %w", err)
				}
			}
		} else {
			if existingUser.Google == nil {
				user.Google.UserID = existingUser.ID
				if err := adapters.UserDbRepository.InsertGoogleUser(ctx, user.Google); err != nil {
					return fmt.Errorf("failed to insert google user: %w", err)
				}
			}
		}

		user, err = adapters.UserDbRepository.GetByEmail(ctx, user.Email)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*User, error) {
	userData := &User{Email: req.Email, Name: req.Name}
	if req.Password != "" {
		if err := userData.Password.Set(req.Password); err != nil {
			return nil, err
		}
	}

	err := s.provider.WithTransaction(func(adapters Adapters) error {
		existingUser, err := adapters.UserDbRepository.GetByEmail(ctx, req.Email)
		if err != nil {
			if errors.Is(err, data.ErrNotFound) && userData.HasPassword() {
				return adapters.UserDbRepository.Insert(ctx, userData)
			}
			return err
		}
		if !existingUser.HasPassword() {
			existingUser.Password = userData.Password
			userData = existingUser
			if err := adapters.UserDbRepository.Update(ctx, userData); err != nil {
				return err
			}
			return nil
		}
		return ErrDuplicateEmail
	})

	if err != nil {
		return nil, err
	}

	return userData, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*User, string, string, error) {
	userData, err := s.userDb.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", "", err
	}

	if len(userData.Password.Hash) > 0 {
		matches, _ := userData.Password.Matches(password)
		if !matches {
			return nil, "", "", ErrInvalidCredentials
		}
	} else {
		return nil, "", "", ErrNoUserPasswordSet
	}

	refreshToken := token.GenerateRefreshToken()
	if err := s.tokenRepo.Save(ctx, strconv.FormatUint(userData.ID, 10), refreshToken); err != nil {
		return nil, "", "", err
	}

	accessToken, _, err := token.GenerateAccessToken(userData.ID)
	if err != nil {
		return nil, "", "", err
	}

	return userData, accessToken, refreshToken, err
}

func (s *Service) ChangePassword(ctx context.Context, newPW string) error {
	claims := authcontext.Claims(ctx)

	id, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		return err
	}

	user := &User{ID: id}
	if err := user.Password.Set(newPW); err != nil {
		return err
	}
	if err := s.userDb.Update(ctx, user); err != nil {
		return err
	}

	return nil
}

func (s *Service) Logout(ctx context.Context, tokenID string) error {
	if tokenID == "" {
		return nil
	}
	return s.tokenRepo.Delete(ctx, tokenID)
}
