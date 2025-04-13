package token

import (
	"context"
	"fmt"
	"strconv"
)

type Service struct {
	tokenRepo TokenRepository
}

func NewService(tokenRepo TokenRepository) *Service {
	return &Service{
		tokenRepo: tokenRepo,
	}
}

func (s *Service) RefreshTokens(ctx context.Context, userID, oldToken string) (string, string, *Claims, error) {
	if err := s.tokenRepo.Valid(ctx, userID, oldToken); err != nil {
		return "", "", nil, err
	}

	newRT := GenerateRefreshToken()
	if err := s.tokenRepo.Save(ctx, userID, newRT); err != nil {
		return "", "", nil, err
	}

	userIDInt, _ := strconv.ParseUint(userID, 10, 64)
	newAT, newClaims, err := GenerateAccessToken(userIDInt)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return newRT, newAT, newClaims, nil
}
