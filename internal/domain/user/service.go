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
	tokenRepo token.TokenRepository
}

func NewService(provider provider, tokenRepo token.TokenRepository) *Service {
	adapters := provider.GetAdapters()
	return &Service{
		provider:  provider,
		userDb:    adapters.UserDbRepository,
		tokenRepo: tokenRepo,
	}
}

func (s *Service) HandleGoogleLogin(ctx context.Context, googleUser *GoogleUser) (UserDTO, error) {
	user := &User{
		Name:   googleUser.Name,
		Email:  googleUser.Email,
		Google: googleUser,
	}

	err := s.provider.WithTransaction(func(adapters Adapters) error {
		fetchedUser, err := adapters.UserDbRepository.GetByEmail(ctx, user.Email)
		if err != nil {
			if errors.Is(err, data.ErrNotFound) {
				if err := adapters.UserDbRepository.Insert(ctx, user); err != nil {
					if !errors.Is(err, data.ErrRecordExists) {
						return fmt.Errorf("failed to insert user: %w", err)
					}
				}
			}
		} else {
			user = fetchedUser
		}
		return nil
	})

	return NewUserDTO(user), err
}

func (s *Service) Register(ctx context.Context, req *RegisterRequest) (UserDTO, error) {
	user := &User{
		Email: req.Email,
		Name:  req.Name,
	}

	if err := user.Password.Set(req.Password); err != nil {
		return UserDTO{}, nil
	}

	err := s.provider.WithTransaction(func(adapters Adapters) error {
		fetched, err := adapters.UserDbRepository.GetByEmail(ctx, user.Email)
		if err != nil {
			if errors.Is(err, data.ErrNotFound) {
				return adapters.UserDbRepository.Insert(ctx, user)
			}
			return err
		}
		if fetched != nil {
			user.ID = fetched.ID
			return adapters.UserDbRepository.Update(ctx, user)
		}
		return ErrDuplicateEmail
	})

	if err != nil {
		return UserDTO{}, nil
	}

	return NewUserDTO(user), nil
}

func (s *Service) Login(ctx context.Context, email, password string) (UserDTO, string, string, error) {
	userData, err := s.userDb.GetByEmail(ctx, email)
	if err != nil {
		return UserDTO{}, "", err
	}

	if len(userData.Password.Hash) > 0 {
		matches, _ := userData.Password.Matches(password)
		if !matches {
			return UserDTO{}, "", ErrInvalidCredentials
		}
	} else {
		return UserDTO{}, "", ErrNoUserPasswordSet
	}

	rt := token.GenerateRefreshToken()
	if err := s.tokenRepo.Store(ctx, rt, strconv.FormatUint(userData.ID, 10)); err != nil {
		return UserDTO{}, "", err
	}

	accessToken, _, err := token.GenerateAccessToken(userData.ID)
	if err != nil {
		return UserDTO{}, "", err
	}

	return NewUserDTO(userData), rt, err
}

func (s *Service) ChangePassword(ctx context.Context, newPW string) error {
	claims := authcontext.GetClaims(ctx)

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
	return s.tokenRepo.Revoke(ctx, tokenID)
}
