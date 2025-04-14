package tinylink_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

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

	return redisClient
}

func TestTinylinkService(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	redisClient := setupRedisDB(t)

	userProvider := user.NewRepositoryProvider(db)
	userAdapters := userProvider.GetAdapters()
	userDb := userAdapters.UserDbRepository
	newUser := &user.User{
		Email: "testuser@gmail.com",
		Name:  "TestUser",
	}
	newUser.Password.Set("test123")
	require.Nil(t, userDb.Insert(ctx, newUser))
	userID := newUser.GetID()

	tlProvider := tinylink.NewRepositoryProvider(db, redisClient)
	tlService := tinylink.NewService(tlProvider)

	// t.Run("Insert test", func(t *testing.T) {
	// 	tlService.Insert(ctx, )
	// })
	// Business logic

	t.Run("Duplicate alias should fail", func(t *testing.T) {
		claims := &token.Claims{UserID: userID}

		req := tinylink.InsertTinylinkRequest{
			OriginalURL: "https://example.com/another",
			Alias:       "cc123",
			Private:     false,
		}

		tl, err := tlService.Insert(ctx, claims, req)
		require.Nil(t, err)
		require.NotEmpty(t, tl)

		tl, err = tlService.Insert(ctx, claims, req)
		require.Nil(t, tl)
		require.NotNil(t, err)
		require.Equal(t, err, tinylink.ErrAliasExists)

		req.Alias = ""
		tl, err = tlService.Insert(ctx, claims, req)
		require.Nil(t, err)
		require.NotNil(t, tl)
		t.Log("TL WITH GENERATED ALIAS: ", tl.Alias)
	})
}

func TestTinylinKSQLite(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	userProvider := user.NewRepositoryProvider(db)
	userAdapters := userProvider.GetAdapters()
	userDb := userAdapters.UserDbRepository
	newUser := &user.User{
		Email: "testuser@gmail.com",
		Name:  "TestUser",
	}
	newUser.Password.Set("test123")
	require.Nil(t, userDb.Insert(ctx, newUser))
	userID := newUser.GetID()

	provider := tinylink.NewRepositoryProvider(db, nil)
	adapters := provider.GetAdapters()
	tlDb := adapters.DBAdapters.TinylinkDBRepository

	t.Run("Create and retrieve tinylink", func(t *testing.T) {
		tl := &tinylink.Tinylink{
			UserID:      &userID,
			OriginalURL: "https://codingchallenges.fyi/challenges/challenge-json-parser/",
			Alias:       "cc123",
			Private:     false,
		}
		require.Nil(t, tlDb.Insert(ctx, tl))
		require.NotZero(t, tl.CreatedAt)
		require.NotEmpty(t, tl.ID)

		tl, err := tlDb.Get(ctx, tl.Alias)
		require.Nil(t, err)
		require.NotEmpty(t, tl)
	})

	var insertedTlID uint64 = 0

	t.Run("List all user tinylinks", func(t *testing.T) {
		tl := &tinylink.Tinylink{
			UserID:      &userID,
			OriginalURL: "https://example.com/another",
			Alias:       "321cc",
			Private:     false,
		}
		err := tlDb.Insert(ctx, tl)
		require.Nil(t, err)
		require.NotZero(t, tl.CreatedAt)
		require.NotEmpty(t, tl.ID)

		links, err := tlDb.List(ctx, userID)
		require.Nil(t, err)
		require.Equal(t, len(links), 2)
		insertedTlID = tl.ID
	})

	t.Run("Update tinylink", func(t *testing.T) {
		tl := &tinylink.Tinylink{
			ID:          insertedTlID,
			UserID:      &userID,
			OriginalURL: "https://example.com/another",
			Alias:       "updateTest123",
			Private:     true,
			ExpiresAt:   time.Now().Add(time.Hour * 1).Unix(),
		}
		require.Nil(t, tlDb.Update(ctx, tl))
		require.True(t, tl.Private)
		require.Equal(t, tl.Version, uint64(1))
		require.NotEmpty(t, tl.ExpiresAt)

		time.Sleep(1 * time.Second)
		oldLV := tl.LastVisited
		require.Nil(t, tlDb.UpdateUsage(ctx, tl.ID))
		require.Equal(t, tl.UsageCount, 1)
		require.NotEqual(t, tl.LastVisited, oldLV)
	})
}
