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
// func TestTinylinkDbRepository_Get(t *testing.T) {}

// func TestTinylinkDbRepository_Create(t *testing.T) {}

// func TestTinylinkRepository_Update(t *testing.T) {}

// func TestTinylinkRepository_Delete(t *testing.T) {}

// Test service
func TestTinylinkService_Create(t *testing.T) {
	redis := setupRedisDB(t)
	db := setupTestDB(t)

	provider := tinylink.NewRepositoryProvider(db, redis)
	tlService := tinylink.NewService(provider)

	mockUrl := "https://www.sometesturl.com/d321ks0dai2lk-321=-321"
	mockAlias := "extra"

	ctx := context.Background()

	sessionID := uuid.NewString()

	req := tinylink.CreateTinylinkRequest{
		URL:   mockUrl,
		Alias: mockAlias,
	}

	// creating public tinylink that will be stored in redis.
	tl, err := tlService.Create(ctx, "", sessionID, req)
	require.NoError(t, err)
	require.NotNil(t, tl)
	require.Equal(t, tl.Alias, req.Alias)
	require.False(t, tl.Private)

	adapters := provider.Adapters()

	// // create mockUser to prevent FOREIGN key error (tinylinks and users related)
	userProvider := user.NewRepositoryProvider(db)
	userRepo := userProvider.Adapters().UserDbRepository
	mockUser := createMockUser(t, ctx, userRepo)

	mockUserID := mockUser.GetID()

	tl, err = tlService.Create(ctx, mockUserID, "", tinylink.CreateTinylinkRequest{
		Alias:   mockAlias,
		URL:     mockUrl,
		Private: false,
	})
	require.Error(t, err)
	require.Nil(t, tl)
	require.Equal(t, err, tinylink.ErrAliasExists)

	// create private
	req.Private = true
	tl, err = tlService.Create(ctx, mockUserID, "", req)
	require.NoError(t, err)
	require.NotNil(t, tl)
	require.Equal(t, tl.Alias, req.Alias)
	require.True(t, tl.Private)

	// validate it
	// ok, err := adapters.TinylinkDBRepository.AliasExistsWithID(ctx, mockUserID, req.Alias)
	// require.NoError(t, err)
	// require.True(t, ok)

	// // compare data
	rowID, url, err := adapters.TinylinkDBRepository.GetPrivateURL(ctx, mockUserID, req.Alias)
	require.NoError(t, err)
	require.Greater(t, rowID, uint64(0))
	require.Equal(t, req.URL, url)

	tl, err = adapters.TinylinkDBRepository.Get(ctx, rowID)
	require.NoError(t, err)
	require.NotNil(t, tl)
	require.Greater(t, tl.CreatedAt, int64(0))
	require.True(t, tl.Private)
}

func TestTinylinkService_Redirect(t *testing.T) {}
