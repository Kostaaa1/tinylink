package token

import (
	"context"
	"fmt"

	"github.com/Kostaaa1/tinylink/internal/domain/auth"
)

type Service struct {
	token Repository
}

func NewService(tokenRepo Repository) *Service {
	return &Service{token: tokenRepo}
}

func (s *Service) ValidateAndRotateTokens(ctx context.Context, userID uint64, oldToken string) (refreshToken string, accessToken string, err error) {
	// validate if provided token matches the stored one in redis
	if err := s.token.Valid(ctx, userID, oldToken); err != nil {
		return "", "", err
	}

	refreshToken = auth.GenerateRefreshToken()
	if err := s.token.Save(ctx, userID, refreshToken, auth.RefreshTokenTTL); err != nil {
		return "", "", err
	}

	accessToken, err = auth.GenerateAccessToken(userID, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	return refreshToken, accessToken, nil
}

type Session struct {
	GuestUUID string
	CSRF      string
	IssuedAt  string
	ExpiresAt string
	IP        string
	UserAgent string
}

func (s *Service) CreateSession(ctx context.Context, sess Session) *Session {
	return nil
}

func (s *Service) GetSession(ctx context.Context, guestUUID string) *Session {
	return nil
}
