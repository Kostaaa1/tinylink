package tinylink_test

import (
	"context"
	"os"
	"testing"

	"github.com/Kostaaa1/tinylink/core/transactor"
	"github.com/Kostaaa1/tinylink/core/transactor/adapters"
	"github.com/Kostaaa1/tinylink/internal/constants"
	"github.com/Kostaaa1/tinylink/internal/domain/auth"
	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	"github.com/Kostaaa1/tinylink/internal/domain/user"
	"github.com/Kostaaa1/tinylink/internal/infra/db"
	"github.com/Kostaaa1/tinylink/internal/infra/postgres"
	"github.com/Kostaaa1/tinylink/internal/infra/redis"
	"github.com/Kostaaa1/tinylink/internal/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func getEnvironment(t *testing.T) (*tinylink.Service, user.Repository) {
	ctx := context.Background()

	pool, err := db.OpenPostgresPool(ctx, os.Getenv("POSTGRES_DSN"))
	require.NoError(t, err)

	redisClient, err := db.OpenRedisConn(ctx, os.Getenv("REDIS_DSN"))
	require.NoError(t, err)

	tx, err := pool.Begin(ctx)
	require.NoError(t, err)

	psqlRepo := postgres.NewTinylinkRepository(tx)
	tlCacheRepo := redis.NewTinylinkRepository(redisClient)

	tlProvider := transactor.NewProvider(psqlRepo, adapters.WithPgxPool(pool))

	t.Cleanup(func() {
		tx.Rollback(ctx)
		redisClient.FlushAll(ctx)
	})

	return tinylink.NewService(tlProvider, tlCacheRepo), postgres.NewUserRepository(tx)
}

func TestTinylinkService_ListUser(t *testing.T) {
	t.Parallel()

	service, userRepo := getEnvironment(t)
	ctx := context.Background()

	userTlCount := 5
	userID := mock.InsertUser(t, userRepo)
	userCtx := auth.UserContext{UserID: &userID}

	t.Run("should create 5 user tinylinks", func(t *testing.T) {
		for i := 0; i < userTlCount; i++ {
			req := mock.CreateTinylinkParams(&userID, nil)
			tl, err := service.Create(ctx, req)
			require.NoError(t, err)
			require.NotNil(t, tl)
			require.Greater(t, tl.ID, uint64(0))
		}
	})

	t.Run("should list 5 user tinylinks", func(t *testing.T) {
		links, err := service.List(ctx, userCtx)
		require.NoError(t, err)
		require.Equal(t, len(links), userTlCount)
	})
}

func TestTinylinkService_ListGuest(t *testing.T) {
	t.Parallel()

	service, _ := getEnvironment(t)
	ctx := context.Background()

	guestTlCount := 8
	guestUUID := uuid.NewString()

	t.Run("should create 5 guest tinylinks", func(t *testing.T) {
		for i := 0; i < guestTlCount; i++ {
			params := mock.CreateTinylinkParams(nil, &guestUUID)
			params.Private = false
			tl, err := service.Create(ctx, params)
			require.NoError(t, err)
			require.NotNil(t, tl)
			require.Greater(t, tl.ID, uint64(0))
		}
	})

	t.Run("should list 8 guest tinylinks", func(t *testing.T) {
		links, err := service.List(ctx, auth.UserContext{GuestUUID: guestUUID})
		require.NoError(t, err)
		require.Equal(t, len(links), guestTlCount)
	})
}

func TestTinylinkService_Create(t *testing.T) {
	t.Parallel()

	service, userRepo := getEnvironment(t)
	ctx := context.Background()

	userID := mock.InsertUser(t, userRepo)
	guestUUID := uuid.NewString()

	t.Run("pass: create private tinylink as user", func(t *testing.T) {
		req := mock.CreateTinylinkParams(&userID, &guestUUID)
		req.Private = true
		tl, err := service.Create(ctx, req)
		require.NoError(t, err)
		require.Greater(t, tl.ID, uint64(0))
		require.Greater(t, tl.CreatedAt.Unix(), int64(0))
	})

	t.Run("fail: create private tinylink as guest", func(t *testing.T) {
		req := mock.CreateTinylinkParams(nil, &guestUUID)
		req.Private = true
		tl, err := service.Create(ctx, req)
		require.Error(t, err)
		require.ErrorIs(t, err, constants.ErrUnauthenticated)
		require.Nil(t, tl)
	})

	t.Run("pass: create public tinylink as guest", func(t *testing.T) {
		req := mock.CreateTinylinkParams(nil, &guestUUID)
		// generate alias with redis
		req.Alias = nil
		req.Private = false
		tl, err := service.Create(ctx, req)
		require.NoError(t, err)
		require.Greater(t, tl.ID, uint64(0))
		require.NotEqual(t, tl.Alias, "")
		require.Greater(t, tl.CreatedAt.Unix(), int64(0))
	})
}

func TestTinylinkService_Update(t *testing.T) {
	// t.Parallel()

	service, userRepo := getEnvironment(t)
	ctx := context.Background()

	userID := mock.InsertUser(t, userRepo)

	var tl *tinylink.Tinylink
	var err error

	t.Run("pass: create private tinylink as user", func(t *testing.T) {
		req := mock.CreateTinylinkParams(&userID, nil)
		req.Private = true
		tl, err = service.Create(ctx, req)
		require.NoError(t, err)
		require.Greater(t, tl.ID, uint64(0))
		require.Greater(t, tl.CreatedAt.Unix(), int64(0))
		require.NotEmpty(t, tl.GuestUUID)
	})

	t.Run("pass: only users (non-guest) can update tinylinks", func(t *testing.T) {
		req := mock.UpdateTinylinkParams(tl.ID, userID)
		tl, err := service.Update(ctx, req)
		require.NoError(t, err)
		require.Greater(t, tl.ID, uint64(0))
		require.Greater(t, tl.CreatedAt.Unix(), int64(0))
		require.Greater(t, tl.UpdatedAt.Unix(), int64(0))
	})

	t.Run("fail: only users (non-guest) can update tinylinks", func(t *testing.T) {
		req := mock.UpdateTinylinkParams(tl.ID, userID)
		tl, err := service.Update(ctx, req)
		require.NoError(t, err)
		require.Greater(t, tl.ID, uint64(0))
		require.Greater(t, tl.CreatedAt.Unix(), int64(0))
		require.Greater(t, tl.UpdatedAt.Unix(), int64(0))
	})
}
