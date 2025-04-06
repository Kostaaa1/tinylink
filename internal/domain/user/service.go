package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kostaaa1/tinylink/internal/common/auth"
	"github.com/Kostaaa1/tinylink/internal/common/data"
)

type Adapters struct {
	UserRepository UserRepository
}

type txProvider interface {
	WithTransaction(txFunc func(adapters Adapters) error) error
	GetDbAdapters() Adapters
}

type Service struct {
	txProvider txProvider
	user       UserRepository
}

func NewService(txProvider txProvider) *Service {
	dbAdapters := txProvider.GetDbAdapters()
	return &Service{
		txProvider: txProvider,
		user:       dbAdapters.UserRepository,
	}
}

func (s *Service) HandleGoogleLogin(ctx context.Context, googleUser *GoogleUser) (UserDTO, error) {
	user := &User{
		Name:   googleUser.Name,
		Email:  googleUser.Email,
		Google: googleUser,
	}

	err := s.txProvider.WithTransaction(func(adapters Adapters) error {
		fetchedUser, err := adapters.UserRepository.GetByEmail(ctx, user.Email)
		if err != nil {
			if errors.Is(err, data.ErrRecordNotFound) {
				if err := adapters.UserRepository.Insert(ctx, user); err != nil {
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

func (s *Service) GetUserFromCtx(ctx context.Context) (UserDTO, error) {
	claims := auth.ClaimsFromCtx(ctx)
	user, err := s.user.GetByEmail(ctx, claims.Email)
	if err != nil {
		return UserDTO{}, err
	}
	return NewUserDTO(user), nil
}

func (s *Service) Register(ctx context.Context, userData *User) (UserDTO, error) {
	err := s.txProvider.WithTransaction(func(adapters Adapters) error {
		fetched, err := adapters.UserRepository.GetByEmail(ctx, userData.Email)
		if err != nil {
			if errors.Is(err, data.ErrRecordNotFound) {
				return adapters.UserRepository.Insert(ctx, userData)
			}
			return err
		}
		if fetched != nil {
			userData.ID = fetched.ID
			return adapters.UserRepository.Update(ctx, userData)
		}
		return ErrDuplicateEmail
	})

	if err != nil {
		return UserDTO{}, nil
	}

	return NewUserDTO(userData), nil
}

func (s *Service) Login(ctx context.Context, email, password string) (UserDTO, error) {
	userData, err := s.user.GetByEmail(ctx, email)
	if err != nil {
		return UserDTO{}, err
	}

	if len(userData.Password.Hash) > 0 {
		matches, err := userData.Password.Matches(password)
		if err != nil {
			return UserDTO{}, err
		}
		if !matches {
			return UserDTO{}, ErrInvalidCredentials
		}
	} else {
		return UserDTO{}, ErrNoUserPasswordSet
	}

	return NewUserDTO(userData), err
}
