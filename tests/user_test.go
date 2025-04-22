package tinylink_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Kostaaa1/tinylink/internal/domain/token"
	"github.com/Kostaaa1/tinylink/internal/domain/user"
	"github.com/stretchr/testify/require"
)

func TestUserService_LoginAfterNormalRegistration(t *testing.T) {
	db := setupTestDB(t)
	redis := setupRedisDB(t)

	provider := user.NewRepositoryProvider(db)
	tokenRepo := token.NewRedisTokenRepository(redis)
	service := user.NewService(provider, tokenRepo)

	ctx := context.Background()
	email := fmt.Sprintf("flow_%d@example.com", time.Now().UnixNano())
	pw := "securepassword123"

	user1, err := service.Register(ctx, user.RegisterRequest{
		Name:     "FlowUser",
		Email:    email,
		Password: pw,
	})
	require.NoError(t, err)
	require.Nil(t, user1.Google)
	require.Equal(t, user1.Email, email)
	require.Greater(t, user1.CreatedAt, int64(0))
	require.NotEmpty(t, user1.Password.Hash)

	user2, err := service.HandleGoogleLogin(ctx, &user.GoogleUser{
		ID:            "333TEST",
		Email:         email,
		VerifiedEmail: true,
		Name:          "Test",
		GivenName:     "Test",
		FamilyName:    "Test",
		Picture:       "https://testimage.png",
	})
	require.NoError(t, err)
	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.CreatedAt, user2.CreatedAt)
	require.NotNil(t, user2.Google)
	require.Greater(t, user2.Google.CreatedAt, int64(0))
	require.NotEmpty(t, user2.Password.Hash)
}

func TestUserService_LoginAfterRegistrationWithGoogle(t *testing.T) {
	db := setupTestDB(t)
	redis := setupRedisDB(t)

	provider := user.NewRepositoryProvider(db)
	tokenRepo := token.NewRedisTokenRepository(redis)
	service := user.NewService(provider, tokenRepo)

	ctx := context.Background()
	email := fmt.Sprintf("flow_%d@example.com", time.Now().UnixNano())
	pw := "securepassword123"

	// Login with google. If user does not exists, it will insert in users and google_users_data tables.
	user1, err := service.HandleGoogleLogin(ctx, &user.GoogleUser{
		Email:         email,
		Name:          "GoogleUserName",
		GivenName:     "GivenName",
		FamilyName:    "FamilyName",
		VerifiedEmail: true,
		Picture:       "https://imgg.png",
		ID:            "TestGoogleID",
	})
	require.NoError(t, err)
	require.Empty(t, user1.Password.Hash)
	require.NotNil(t, user1.ID)
	require.NotNil(t, user1.Google)
	require.Greater(t, user1.CreatedAt, int64(0))
	require.Greater(t, user1.Google.CreatedAt, int64(0))

	// Trying to login with password, but password is not set because user was created upon google login.
	loggedUser, newAt, newRt, err := service.Login(ctx, email, pw)
	require.Error(t, err)
	require.Nil(t, loggedUser)
	require.Empty(t, newAt)
	require.Empty(t, newRt)
	require.Equal(t, err, user.ErrNoUserPasswordSet)

	// Registering with password
	user2, err := service.Register(ctx, user.RegisterRequest{
		Email:    user1.Email,
		Name:     user1.Name,
		Password: pw,
	})
	require.NoError(t, err)
	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.Name, user2.Name)
	require.NotEmpty(t, user2.Password.Hash)

	// Login with invalid password
	loggedUser, newAt, newRt, err = service.Login(ctx, email, "SomeRandomPassword")
	require.Error(t, err)
	require.Equal(t, err, user.ErrInvalidCredentials)
	require.Nil(t, loggedUser)
	require.Empty(t, newAt)
	require.Empty(t, newRt)

	// login with correct password
	loggedUser, newAt, newRt, err = service.Login(ctx, email, pw)
	require.NoError(t, err)
	require.NotNil(t, loggedUser)
	require.NotNil(t, loggedUser.Google)
	require.NotEmpty(t, newAt)
	require.NotEmpty(t, newRt)
	require.Greater(t, loggedUser.CreatedAt, int64(0))
	require.Greater(t, loggedUser.Google.CreatedAt, int64(0))
}

func TestUserService_RegistrationFlow(t *testing.T) {
	db := setupTestDB(t)
	provider := user.NewRepositoryProvider(db)
	service := user.NewService(provider, nil)

	ctx := context.Background()
	email := fmt.Sprintf("flow_%d@example.com", time.Now().UnixNano())
	pw := "secure123"

	// Step 1: Register with password (should insert new user)
	user2, err := service.Register(ctx, user.RegisterRequest{
		Name:     "FlowUser",
		Email:    email,
		Password: pw,
	})
	require.NoError(t, err)
	require.Equal(t, user2.Email, email)
	require.NotEmpty(t, user2.Password.Hash)

	// Step 2: Attempt register with same email (should fail)
	_, err = service.Register(ctx, user.RegisterRequest{
		Name:     "FlowUser",
		Email:    email,
		Password: "anotherpass",
	})
	require.ErrorIs(t, err, user.ErrDuplicateEmail)
}
