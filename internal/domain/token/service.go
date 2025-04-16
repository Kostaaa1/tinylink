package token

import (
	"context"
	"fmt"
	"strconv"
)

type Service struct {
	tokenRepo Repository
}

func NewService(tokenRepo Repository) *Service {
	return &Service{
		tokenRepo: tokenRepo,
	}
}

func (s *Service) RefreshTokens(ctx context.Context, userID, oldToken string) (string, string, Claims, error) {
	if err := s.tokenRepo.Valid(ctx, userID, oldToken); err != nil {
		return "", "", Claims{}, err
	}

	newRT := GenerateRefreshToken()
	if err := s.tokenRepo.Save(ctx, userID, newRT); err != nil {
		return "", "", Claims{}, err
	}

	userIDInt, _ := strconv.ParseUint(userID, 10, 64)
	newAT, newClaims, err := GenerateAccessToken(userIDInt)
	if err != nil {
		return "", "", Claims{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	return newRT, newAT, newClaims, nil
}
