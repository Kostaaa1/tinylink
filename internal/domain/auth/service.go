package auth

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

func (s *Service) RefreshTokens(ctx context.Context, oldRefreshToken, userID string) (string, string, *Claims, error) {
	newRT := GenerateRefreshToken()
	userID, err := s.tokenRepo.TxDelOldAndInsertNew(ctx, userID, oldRefreshToken, newRT, RefreshTokenDuration)
	if err != nil {
		return "", "", nil, err
	}

	userIDInt, _ := strconv.ParseUint(userID, 10, 64)
	newAT, newClaims, err := GenerateAccessToken(userIDInt)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return newRT, newAT, newClaims, nil
}
