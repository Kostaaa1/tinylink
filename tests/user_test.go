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

func TestUserService_LoginAfterRegistrationWithGoogle(t *testing.T) {
	db := setupTestDB(t)
	redis := setupRedisDB(t)

	provider := user.NewRepositoryProvider(db)
	tokenRepo := token.NewRedisTokenRepository(redis)
	service := user.NewService(provider, tokenRepo)

	ctx := context.Background()
	email := fmt.Sprintf("flow_%d@example.com", time.Now().UnixNano())
	pw := "securepassword123"

	// Step 1: First login with Google. It will create user if it does not exist. It will update user if exists
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
	require.NotNil(t, user1.Google)
	require.Greater(t, user1.CreatedAt, int64(0))
	require.Greater(t, user1.Google.CreatedAt, int64(0))

	loggedUser, newAt, newRt, err := service.Login(ctx, email, pw)
	require.Error(t, err)
	require.Nil(t, loggedUser)
	require.Empty(t, newAt)
	require.Empty(t, newRt)
	require.Equal(t, err, user.ErrNoUserPasswordSet)

	user2, err := service.Register(ctx, user.RegisterRequest{
		Email:    user1.Email,
		Name:     user1.Name,
		Password: pw,
	})
	require.NoError(t, err)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.Name, user2.Name)
	require.NotEmpty(t, user2.Password.Hash)

	loggedUser, newAt, newRt, err = service.Login(ctx, email, "SomeRandomPassword")
	require.Error(t, err)
	require.Equal(t, err, user.ErrInvalidCredentials)
	require.Nil(t, loggedUser)
	require.Empty(t, newAt)
	require.Empty(t, newRt)

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

	// Step 1: Register without password
	user1, err := service.Register(ctx, user.RegisterRequest{
		Name:  "FlowUser",
		Email: email,
	})
	require.NoError(t, err)
	require.Empty(t, user1.Password.Hash)

	// Step 2: Register again with password (should update)
	user2, err := service.Register(ctx, user.RegisterRequest{
		Name:     "FlowUser",
		Email:    email,
		Password: "secure123",
	})
	require.NoError(t, err)
	require.Equal(t, user2.Email, email)
	require.NotEmpty(t, user2.Password.Hash)

	// Step 3: Attempt third register (should fail)
	_, err = service.Register(ctx, user.RegisterRequest{
		Name:     "FlowUser",
		Email:    email,
		Password: "anotherpass",
	})
	require.ErrorIs(t, err, user.ErrDuplicateEmail)
}

// func TestUser(t *testing.T) {
// 	err := exec.Command("rm", "-rf", "./tinylink.db").Run()
// 	require.Nil(t, err)

// 	conf := config.SQLConfig{
// 		SQLitePath:   "./tinylink.db",
// 		MaxOpenConns: 25,
// 		MaxIdleConns: 25,
// 	}

// 	db, err := sqlitedb.StartDB(conf)
// 	require.Nil(t, err)

// 	provider := user.NewRepositoryProvider(db)
// 	userService := user.NewService(provider)

// 	ctx := context.Background()

// 	gUser := &user.GoogleUser{
// 		ID:            "33333",
// 		Email:         "kostaarsic123@gmail.com",
// 		Name:          "Kosta",
// 		GivenName:     "Arsic",
// 		FamilyName:    "Kosta Arsic",
// 		Picture:       "testpicture-url.com",
// 		VerifiedEmail: true,
// 	}

// 	userDTO, err := userService.HandleGoogleLogin(ctx, gUser)
// 	require.Nil(t, err)
// 	require.NotZero(t, userDTO.CreatedAt)
// 	require.NotNil(t, userDTO)

// 	newUser := &user.User{
// 		Email: userDTO.Email,
// 		Name:  userDTO.Name,
// 	}
// 	err = newUser.Password.Set("TestPassword123")
// 	require.Nil(t, err)

// 	err = userService.Register(ctx, newUser)
// 	require.Nil(t, err)

// 	userRepo := provider.GetDbAdapters().UserRepository
// 	fetchedUser, err := userRepo.GetByEmail(ctx, userDTO.Email)
// 	require.Nil(t, err)

// 	require.NotEmpty(t, fetchedUser)
// 	require.NotNil(t, fetchedUser)

// 	t.Log("logovan: ", fetchedUser)
// }
