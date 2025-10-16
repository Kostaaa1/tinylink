package user

import (
	"context"

	"github.com/Kostaaa1/tinylink/internal/domain/auth"
	"github.com/Kostaaa1/tinylink/internal/domain/token"
)

type Service struct {
	user  Repository
	token token.Repository
}

func NewService(user Repository, token token.Repository) *Service {
	return &Service{
		user:  user,
		token: token,
	}
}

func (s *Service) HandleGoogleLogin(ctx context.Context, googleUser *GoogleUser) (*User, error) {
	user := &User{
		Name:   googleUser.Name,
		Email:  googleUser.Email,
		Google: googleUser,
	}

	// err := s.txManager.WithTx(ctx, func(ctx context.Context) error {
	// 	existingUser, err := s.user.GetByEmail(ctx, user.Email)
	// 	if err != nil && errors.Is(err, constants.ErrNotFound) {
	// 		if err := s.user.Insert(ctx, user); err != nil {
	// 			return fmt.Errorf("failed to insert user: %w", err)
	// 		}
	// 	} else {
	// 		if existingUser.Google == nil {
	// 			user.Google.UserID = existingUser.ID
	// 			if err := s.user.InsertGoogleUser(ctx, user.Google); err != nil {
	// 				return fmt.Errorf("failed to insert google user: %w", err)
	// 			}
	// 		}
	// 	}

	// 	user, err = s.user.GetByEmail(ctx, user.Email)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	return nil
	// })

	// if err != nil {
	// 	return nil, err
	// }

	return user, nil
}

func (s *Service) Register(ctx context.Context, userData *User) error {
	return s.user.Insert(ctx, userData)
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
