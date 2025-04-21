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

	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	"github.com/Kostaaa1/tinylink/internal/domain/user"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/db/sqlitedb"
	"github.com/Kostaaa1/tinylink/pkg/config"
	"github.com/google/uuid"
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

// Test repository
// func TestTinylinkDbRepository_Create(t *testing.T) {}
// func TestTinylinkRepository_Update(t *testing.T) {}
// func TestTinylinkRepository_Delete(t *testing.T) {}

// Test service
func TestTinylinkService_Create(t *testing.T) {
	redis := setupRedisDB(t)
	db := setupTestDB(t)

	provider := tinylink.NewRepositoryProvider(db, redis)
	tlService := tinylink.NewService(provider)

	mockUrl := "https://www.youtube.com/watch?v=o8NPllzkFhE&t=11s"
	mockAlias := "extra"

	sessionID := uuid.NewString()
	req := tinylink.InsertTinylinkRequest{
		URL:       mockUrl,
		Alias:     mockAlias,
		SessionID: sessionID,
		Private:   true,
	}

	ctx := context.Background()

	tl, err := tlService.Create(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, tl)
	require.Equal(t, tl.Alias, req.Alias)
	require.False(t, tl.Private)

	val, err := redis.Exists(ctx, fmt.Sprintf("%s:%s", sessionID, mockAlias)).Result()
	require.NoError(t, err)
	require.Greater(t, val, int64(0))
}

func TestTinylinkService_Redirect(t *testing.T) {}
