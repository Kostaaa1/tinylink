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
	GetAdapters() Adapters
}

type Service struct {
	txProvider txProvider
	user       UserRepository
}

func NewService(txProvider txProvider) *Service {
	dbAdapters := txProvider.GetAdapters()
	fmt.Println("DB Adapters: ", dbAdapters)
	return &Service{
		txProvider: txProvider,
		user:       dbAdapters.UserRepository,
	}
}

func (s *Service) HandleGoogleLogin(ctx context.Context, googleUser *GoogleUser) (UserDTO, error) {
	var err error
	user := new(User)

	err = s.txProvider.WithTransaction(func(adapters Adapters) error {
		user, err = adapters.UserRepository.GetByEmail(ctx, googleUser.Email)
		if err != nil {
			return err
		}
		userID := strconv.FormatUint(user.ID, 10)

		if err := adapters.UserRepository.Delete(ctx, userID); err != nil {
			return err
		}

		return errors.New("forced error ot test rollback")
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

func (s *Service) Register(ctx context.Context, userData *User) error {
	_, err := s.user.GetByEmail(ctx, userData.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return s.user.Insert(ctx, userData)
		default:
			return err
		}
	}
	return ErrDuplicateEmail
}

func (s *Service) Login(ctx context.Context, email, password string) (UserDTO, error) {
	userData, err := s.user.GetByEmail(ctx, email)
	if err != nil {
		return UserDTO{}, err
	}
	matches, err := userData.Password.Matches(password)
	if err != nil {
		return UserDTO{}, err
	}
	if !matches {
		return UserDTO{}, ErrInvalidCredentials
	}
	return NewUserDTO(userData), err
}
