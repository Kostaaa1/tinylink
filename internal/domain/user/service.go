package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kostaaa1/tinylink/core/transactor"
	"github.com/Kostaaa1/tinylink/internal/constants"
	"github.com/Kostaaa1/tinylink/internal/domain/auth"
	"github.com/Kostaaa1/tinylink/internal/domain/token"
)

type Service struct {
	user     Repository
	token    token.Repository
	provider *transactor.Provider[Repository]
}

func NewService(user Repository, token token.Repository, provider *transactor.Provider[Repository]) *Service {
	return &Service{
		user:     user,
		token:    token,
		provider: provider,
	}
}

func (s *Service) HandleGoogleLogin(ctx context.Context, googleUser *GoogleUser) (*User, error) {
	user := &User{
		Name:   googleUser.Name,
		Email:  googleUser.Email,
		Google: googleUser,
	}

	// needs to upsert, create upsert repo method and just call that instead
	err := s.provider.WithTx(ctx, func(repo Repository) error {
		existingUser, err := repo.GetByEmail(ctx, user.Email)
		if err != nil {
			if errors.Is(err, constants.ErrNotFound) {
				if err := repo.Insert(ctx, user); err != nil {
					return fmt.Errorf("failed to insert user: %w", err)
				}
			}
		} else {
			if existingUser.Google == nil {
				user.Google.UserID = existingUser.ID
				if err := repo.InsertGoogleUser(ctx, user.Google); err != nil {
					return fmt.Errorf("failed to insert google user: %w", err)
				}
			}
		}

		user, err = repo.GetByEmail(ctx, user.Email)
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

func (s *Service) Register(ctx context.Context, userData *User) error {
	return s.user.Insert(ctx, userData)

	// userData := &User{Email: req.Email, Name: req.Name}
	// if err := userData.Password.Set(req.Password); err != nil {
	// 	return nil, err
	// }
	// err := s.user.WithTransaction(ctx, func(repos *UserSQLiteRepository) error {
	// return s.provider.WithTx(ctx, func(repos *store.PostgresRepository) error {
	// 	existingUser, err := repos.GetByEmail(ctx, userData.Email)
	// 	if err != nil {
	// 		if errors.Is(err, constants.ErrNotFound) && userData.HasPassword() {
	// 			return repos.Insert(ctx, userData)
	// 		}
	// 		return err
	// 	}
	// 	if !existingUser.HasPassword() {
	// 		existingUser.Password = userData.Password
	// 		userData = existingUser
	// 		if err := repos.Update(ctx, userData); err != nil {
	// 			return err
	// 		}
	// 		return nil
	// 	}
	// 	return nil
	// })
}

func (s *Service) Login(ctx context.Context, email, password string) (*User, string, string, error) {
	userData, err := s.user.GetByEmail(ctx, email)
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

	refreshToken := auth.GenerateRefreshToken()
	if err := s.token.Save(ctx, userData.ID, refreshToken, auth.RefreshTokenTTL); err != nil {
		return nil, "", "", err
	}

	accessToken, err := auth.GenerateAccessToken(userData.ID, nil)
	if err != nil {
		return nil, "", "", err
	}

	return userData, accessToken, refreshToken, err
}

func (s *Service) ChangePassword(ctx context.Context, userID uint64, oldPW, newPW string) error {
	// userData := &user.User{ID: userID}
	// if err := userData.Password.Set(newPW); err != nil {
	// 	return err
	// }

	userData, err := s.user.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	matches, err := userData.Password.Matches(oldPW)
	if err != nil {
		return err
	}

	if !matches {
		return ErrInvalidCredentials
	}

	if err := userData.Password.Set(newPW); err != nil {
		return err
	}

	if err := s.user.Update(ctx, userData); err != nil {
		return err
	}

	return nil
}

func (s *Service) Logout(ctx context.Context, userID uint64) error {
	return s.token.Revoke(ctx, userID)
}
