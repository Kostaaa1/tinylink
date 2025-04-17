package tinylink_test

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	"github.com/Kostaaa1/tinylink/internal/domain/token"
	"github.com/Kostaaa1/tinylink/internal/domain/user"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/db/sqlitedb"
	"github.com/Kostaaa1/tinylink/pkg/config"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "tinylink_test.db")

	conf := config.SQLConfig{
		SQLitePath:   dbPath,
		MaxOpenConns: 25,
		MaxIdleConns: 25,
	}

	db, err := sqlitedb.StartDB(conf)
	require.NoError(t, err)

	file, err := os.ReadFile("../sql/tables.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(file))
	require.NoError(t, err)

	return db
}

func setupRedisDB(t *testing.T) *redis.Client {
	t.Helper()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "lagaosiprovidnokopas",
		DB:       0,
		PoolSize: 25,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.Nil(t, redisClient.Ping(ctx).Err())
	redisClient.FlushAll(ctx)

	return redisClient
}

func createMockUser(t *testing.T, ctx context.Context, userDb user.UserRepository) *user.User {
	mockUser := &user.User{
		Email: fmt.Sprintf("testuser%d@gmail.com", rand.Intn(100)),
		Name:  "TestUser",
	}
	mockUser.Password.Set("test123")
	require.Nil(t, userDb.Insert(ctx, mockUser))
	return mockUser
}

func TestTinylinkRepository_Insert(t *testing.T) {
	db := setupTestDB(t)
	redis := setupRedisDB(t)
	provider := tinylink.NewRepositoryProvider(db, redis)
	tlService := tinylink.NewService(provider)

	ctx := context.Background()

	userProvider := user.NewRepositoryProvider(db)
	userDb := userProvider.GetAdapters().UserDbRepository
	user1 := createMockUser(t, ctx, userDb)

	mockUrl := "https://medium.com/nerd-for-tech/redis-getting-notified-when-a-key-is-expired-or-changed-ca3e1f1c7f0a"

	emptyClaims := token.Claims{}
	tl1, err := tlService.Insert(ctx, emptyClaims, tinylink.InsertTinylinkRequest{
		URL:     mockUrl,
		Alias:   "extra",
		Private: true,
	})
	require.NoError(t, err)
	require.NotNil(t, tl1)
	require.False(t, tl1.Private) // no user id, so it will automatically be public

	claims := token.Claims{UserID: user1.GetID()}

	req := &tinylink.InsertTinylinkRequest{
		URL:   mockUrl,
		Alias: "extra",
	}
	tl2, err := tlService.Insert(ctx, claims, *req)
	require.Error(t, err)
	require.Equal(t, err, tinylink.ErrAliasExists)
	require.Nil(t, tl2)

	req.Private = true
	tl3, err := tlService.Insert(ctx, claims, *req)
	require.NoError(t, err)
	require.NotNil(t, tl3)
	require.True(t, tl3.Private)
	require.Equal(t, tl3.URL, mockUrl)
	require.Equal(t, tl3.Alias, "extra")
	require.Greater(t, tl3.CreatedAt, int64(0))

	req.Private = true
	tl4, err := tlService.Insert(ctx, claims, *req)
	require.Error(t, err)
	require.Equal(t, err, tinylink.ErrAliasExists)
	require.Nil(t, tl4)

	// tl5, err := tlService.Update(ctx, claims, tinylink.UpdateTinylinkRequest{
	// 	ID:      tl3.ID,
	// 	Private: false,
	// })
	// require.NoError(t, err)
	// require.NotNil(t, tl5)
	// require.False(t, tl5.Private)
	// require.NotEmpty(t, tl5.Alias)
	// require.NotEmpty(t, tl5.URL)
}

func TestTinylinkRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	redis := setupRedisDB(t)
	provider := tinylink.NewRepositoryProvider(db, redis)
	tlService := tinylink.NewService(provider)

	ctx := context.Background()

	userProvider := user.NewRepositoryProvider(db)
	userDb := userProvider.GetAdapters().UserDbRepository
	user1 := createMockUser(t, ctx, userDb)
	user2 := createMockUser(t, ctx, userDb)

	mockUrl := "https://medium.com/nerd-for-tech/redis-getting-notified-when-a-key-is-expired-or-changed-ca3e1f1c7f0a"
	mockAlias := "extra"

	tl, err := tlService.Insert(ctx, token.Claims{UserID: user1.GetID()}, tinylink.InsertTinylinkRequest{
		URL:     mockUrl,
		Alias:   mockAlias,
		Private: true,
	})
	require.NoError(t, err)
	require.NotNil(t, tl)

	tl2, err := tlService.Update(ctx, token.Claims{UserID: user2.GetID()}, tinylink.UpdateTinylinkRequest{
		URL:     &mockUrl,
		Alias:   &mockAlias,
		Private: false,
	})
	require.Error(t, err)
	require.Equal(t, err, data.ErrNotFound)
	require.Nil(t, tl2)
}

func TestTinylinkRepository_Delete(t *testing.T) {}

// func TestTinylinkService(t *testing.T) {
// 	ctx := context.Background()
// 	db := setupTestDB(t)
// 	redisClient := setupRedisDB(t)

// 	userProvider := user.NewRepositoryProvider(db)
// 	userAdapters := userProvider.GetAdapters()
// 	userDb := userAdapters.UserDbRepository

// 	mockUser := createMockUser(t, ctx, userDb)

// 	tlProvider := tinylink.NewRepositoryProvider(db, redisClient)
// 	tlService := tinylink.NewService(tlProvider)

// 	t.Run("Anonymous insert (no user) - redis", func(t *testing.T) {
// 		req := tinylink.InsertTinylinkRequest{
// 			URL:     "https://example.com/another",
// 			Alias:   "cc123",
// 			Private: true,
// 		}
// 		tl, err := tlService.Insert(ctx, token.Claims{}, req)
// 		require.Nil(t, err)
// 		require.NotNil(t, tl)
// 		require.False(t, tl.Private)
// 	})

// 	t.Run("User insert - db", func(t *testing.T) {
// 		claims := token.Claims{UserID: mockUser.GetID()}
// 		req := tinylink.InsertTinylinkRequest{
// 			URL:     "https://example.com/another",
// 			Alias:   "cc123",
// 			Private: true,
// 		}
// 		tl, err := tlService.Insert(ctx, claims, req)
// 		require.Nil(t, err)
// 		require.NotNil(t, tl)
// 		require.True(t, tl.Private)
// 	})

// 	t.Run("User insert - should fail because private is already created", func(t *testing.T) {
// 		claims := token.Claims{UserID: mockUser.GetID()}
// 		req := tinylink.InsertTinylinkRequest{
// 			URL:     "https://example.com/another",
// 			Alias:   "cc123",
// 			Private: true,
// 		}
// 		tl, err := tlService.Insert(ctx, claims, req)
// 		require.NotNil(t, err)
// 		require.Equal(t, err, tinylink.ErrAliasExists)
// 		require.Nil(t, tl)
// 	})
// }

// func TestTinylinKSQLite(t *testing.T) {
// 	db := setupTestDB(t)
// 	ctx := context.Background()

// 	userProvider := user.NewRepositoryProvider(db)
// 	userAdapters := userProvider.GetAdapters()
// 	userDb := userAdapters.UserDbRepository
// 	newUser := &user.User{
// 		Email: "testuser@gmail.com",
// 		Name:  "TestUser",
// 	}
// 	newUser.Password.Set("test123")
// 	require.Nil(t, userDb.Insert(ctx, newUser))
// 	userID := newUser.GetID()

// 	provider := tinylink.NewRepositoryProvider(db, nil)
// 	adapters := provider.GetAdapters()
// 	tlDb := adapters.DBAdapters.TinylinkDBRepository

// 	t.Run("Create and retrieve tinylink", func(t *testing.T) {
// 		tl := &tinylink.Tinylink{
// 			UserID:  &userID,
// 			URL:     "https://codingchallenges.fyi/challenges/challenge-json-parser/",
// 			Alias:   "cc123",
// 			Private: false,
// 		}
// 		require.Nil(t, tlDb.Insert(ctx, tl))
// 		require.NotZero(t, tl.CreatedAt)
// 		require.NotEmpty(t, tl.ID)

// 		tl, err := tlDb.Get(ctx, tl.Alias)
// 		require.Nil(t, err)
// 		require.NotEmpty(t, tl)
// 	})

// 	var insertedTlID uint64 = 0

// 	t.Run("List all user tinylinks", func(t *testing.T) {
// 		tl := &tinylink.Tinylink{
// 			UserID:  &userID,
// 			URL:     "https://example.com/another",
// 			Alias:   "321cc",
// 			Private: false,
// 		}
// 		err := tlDb.Insert(ctx, tl)
// 		require.Nil(t, err)
// 		require.NotZero(t, tl.CreatedAt)
// 		require.NotEmpty(t, tl.ID)
// 		links, err := tlDb.List(ctx, userID)
// 		require.Nil(t, err)
// 		require.Equal(t, len(links), 2)
// 		insertedTlID = tl.ID
// 	})

// 	t.Run("Update tinylink", func(t *testing.T) {
// 		tl := &tinylink.Tinylink{
// 			ID:        insertedTlID,
// 			UserID:    &userID,
// 			URL:       "https://example.com/another",
// 			Alias:     "updateTest123",
// 			Private:   true,
// 			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
// 		}
// 		require.Nil(t, tlDb.Update(ctx, tl))
// 		require.True(t, tl.Private)
// 		require.Equal(t, tl.Version, uint64(1))
// 		require.NotEmpty(t, tl.ExpiresAt)
// 		time.Sleep(1 * time.Second)
// 		oldLV := tl.LastVisited
// 		require.Nil(t, tlDb.UpdateUsage(ctx, tl.ID))
// 		require.Equal(t, tl.UsageCount, 1)
// 		require.NotEqual(t, tl.LastVisited, oldLV)
// 	})
// }
