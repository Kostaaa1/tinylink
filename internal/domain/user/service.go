package user

import (
	"context"
	"errors"
	"fmt"
	"strconv"

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
	return &Service{
		txProvider: txProvider,
		user:       txProvider.GetDbAdapters().UserRepository,
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

func (s *Service) Register(ctx context.Context, req *RegisterRequest) (UserDTO, error) {
	user := &User{
		Email: req.Email,
		Name:  req.Name,
	}

	if err := user.Password.Set(req.Password); err != nil {
		return UserDTO{}, nil
	}

	err := s.txProvider.WithTransaction(func(adapters Adapters) error {
		fetched, err := adapters.UserRepository.GetByEmail(ctx, user.Email)
		if err != nil {
			if errors.Is(err, data.ErrRecordNotFound) {
				return adapters.UserRepository.Insert(ctx, user)
			}
			return err
		}
		if fetched != nil {
			user.ID = fetched.ID
			return adapters.UserRepository.Update(ctx, user)
		}
		return ErrDuplicateEmail
	})

	if err != nil {
		return UserDTO{}, nil
	}

	return NewUserDTO(user), nil
}

func (s *Service) Login(ctx context.Context, email, password string) (UserDTO, error) {
	userData, err := s.user.GetByEmail(ctx, email)
	if err != nil {
		return UserDTO{}, err
	}

	if len(userData.Password.Hash) > 0 {
		matches, _ := userData.Password.Matches(password)
		if !matches {
			return UserDTO{}, ErrInvalidCredentials
		}
	} else {
		return UserDTO{}, ErrNoUserPasswordSet
	}

	return NewUserDTO(userData), err
}

func (s *Service) ChangePassword(ctx context.Context, newPW string) error {
	claims := auth.ClaimsFromCtx(ctx)

	id, err := strconv.ParseUint(claims.ID, 10, 64)
	if err != nil {
		return err
	}

	user := &User{Email: claims.Email, ID: id}
	if err := user.Password.Set(newPW); err != nil {
		return err
	}
	if err := s.user.Update(ctx, user); err != nil {
		return err
	}

	return nil
}
