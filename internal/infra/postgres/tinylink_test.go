package postgres

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"testing"

// 	"github.com/Kostaaa1/tinylink/internal/constants"
// 	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
// 	"github.com/Kostaaa1/tinylink/internal/mock"
// 	"github.com/google/uuid"

// 	"github.com/jackc/pgx/v5/pgxpool"
// 	"github.com/joho/godotenv"
// 	"github.com/stretchr/testify/require"
// )

// var (
// 	ctx = context.Background()
// )

// func init() {
// 	godotenv.Load("../../../.env")
// }

// func setupPSQLTest(t *testing.T) (*TinylinkRepository, *UserRepository) {
// 	conn, err := pgxpool.New(context.Background(), os.Getenv("POSTGRES_DSN"))
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
// 		os.Exit(1)
// 	}

// 	tx, err := conn.Begin(ctx)
// 	require.NoError(t, err)

// 	t.Cleanup(func() {
// 		_ = tx.Rollback(ctx)
// 		conn.Close()
// 	})

// 	return &TinylinkRepository{db: tx}, NewUserRepository(tx)
// }

// func TestTinylinkRepository_List(t *testing.T) {
// 	t.Parallel()

// 	repo, userRepo := setupPSQLTest(t)
// 	userID := mock.InsertUser(t, userRepo)

// 	t.Run("creates 3 user tinylinks and list them", func(t *testing.T) {
// 		count := 3
// 		for i := 0; i < count; i++ {
// 			tl := mock.UserTinylink(userID)
// 			err := repo.Insert(ctx, tl)
// 			require.NoError(t, err)
// 			require.NotNil(t, tl)
// 		}
// 		tinylinks, err := repo.ListByUserID(ctx, userID)
// 		require.NoError(t, err)
// 		require.NotNil(t, tinylinks)
// 		require.Equal(t, len(tinylinks), count)
// 	})

// 	t.Run("creates 3 guest tinylinks and list them", func(t *testing.T) {
// 		count := 3
// 		guestID := uuid.NewString()
// 		for i := 0; i < count; i++ {
// 			tl := mock.GuestTinylink(guestID)
// 			err := repo.Insert(ctx, tl)
// 			require.NoError(t, err)
// 			require.NotNil(t, tl)
// 		}
// 		tinylinks, err := repo.ListByGuestUUID(ctx, guestID)
// 		require.NoError(t, err)
// 		require.NotNil(t, tinylinks)
// 		require.Equal(t, len(tinylinks), count)
// 	})
// }

// func assertTinylinkPersisted(t *testing.T, repo *PostgresRepository, tl *tinylink.Tinylink) {
// 	t.Helper()

// 	err := repo.Insert(ctx, tl)
// 	require.NoError(t, err)
// 	require.NotNil(t, tl)
// 	require.Greater(t, tl.ID, uint64(0))
// 	require.Equal(t, tl.Version, uint64(1))
// 	require.Greater(t, tl.CreatedAt.Unix(), int64(0))
// 	require.Zero(t, tl.UpdatedAt)

// 	got, err := repo.Get(ctx, tl.ID)
// 	require.NoError(t, err)
// 	require.NotNil(t, got)
// 	require.Equal(t, tl.ID, got.ID)
// 	require.Equal(t, tl.Alias, got.Alias)
// 	require.Equal(t, tl.URL, got.URL)
// 	require.Equal(t, tl.UserID, got.UserID)
// 	require.Equal(t, tl.CreatedAt, got.CreatedAt)
// }

// func TestTinylinkRepository_GetUserTinylink(t *testing.T) {
// 	t.Parallel()
// 	repo, userRepo := setupPSQLTest(t)
// 	userID := mock.InsertUser(t, userRepo)
// 	tl := mock.UserTinylink(userID)
// 	assertTinylinkPersisted(t, repo, tl)
// }

// func TestTinylinkRepository_GetguestTinylink(t *testing.T) {
// 	t.Parallel()
// 	repo, _ := setupPSQLTest(t)
// 	tl := mock.GuestTinylink(uuid.NewString())
// 	assertTinylinkPersisted(t, repo, tl)
// }

// func TestTinylinkRepository_Upsert(t *testing.T) {
// 	t.Parallel()

// 	repo, userRepo := setupPSQLTest(t)
// 	userID := mock.InsertUser(t, userRepo)

// 	tl := mock.UserTinylink(userID)
// 	tl.Private = true

// 	err := repo.Insert(ctx, tl)
// 	require.NoError(t, err)
// 	require.NotNil(t, tl)
// 	require.Greater(t, tl.ID, uint64(0))
// 	require.Equal(t, tl.Version, uint64(1))
// 	require.True(t, tl.Private)
// 	require.NotEmpty(t, tl.Domain)
// 	require.Greater(t, tl.CreatedAt.Unix(), int64(0))
// 	require.Zero(t, tl.UpdatedAt)

// 	newAlias := "updated_testalias"
// 	newDomain := "updated.yoo.com"
// 	tl.Private = false
// 	tl.Alias = newAlias
// 	tl.Domain = newDomain

// 	err = repo.Update(ctx, tl)
// 	require.NoError(t, err)
// 	require.NotNil(t, tl)
// 	require.Greater(t, tl.ID, uint64(0))
// 	require.Equal(t, tl.Alias, newAlias)
// 	require.Equal(t, tl.Domain, newDomain)
// 	require.Equal(t, tl.Version, uint64(2))
// 	require.False(t, tl.Private)
// 	require.Greater(t, tl.UpdatedAt.Unix(), int64(0))
// }

// func TestTinylinkRepository_ErrorAliasExist(t *testing.T) {
// 	t.Parallel()

// 	repo, userRepo := setupPSQLTest(t)
// 	userID := mock.InsertUser(t, userRepo)

// 	tl1 := mock.UserTinylink(userID)
// 	err := repo.Insert(ctx, tl1)
// 	require.NoError(t, err)

// 	tl2 := mock.UserTinylink(userID)
// 	tl2.Alias = tl1.Alias
// 	err = repo.Insert(ctx, tl2)
// 	require.Error(t, err)
// 	require.Equal(t, err, tinylink.ErrAliasExists)

// 	// But if different user tries to create public tinylink with same private alias
// 	// should manually rollback because the error occurred. i am getting error:
// 	// ERROR: current transaction is aborted, commands ignored until end of transaction
// 	// tl2.UserID = 456
// 	// err = repo.Insert(ctx, tl2)
// 	// require.NoError(t, err)
// 	// require.Greater(t, tl2.CreatedAt.Unix(), int64(0))
// }

// func TestTinylinkRepository_Delete(t *testing.T) {
// 	t.Parallel()
// 	repo, userRepo := setupPSQLTest(t)

// 	userID := mock.InsertUser(t, userRepo)
// 	tl := mock.UserTinylink(userID)
// 	tl.Private = true

// 	err := repo.Insert(ctx, tl)
// 	require.NoError(t, err)
// 	require.NotNil(t, tl)
// 	require.GreaterOrEqual(t, tl.ID, uint64(0))

// 	_, err = repo.Get(ctx, tl.ID)
// 	require.NoError(t, err)

// 	err = repo.Delete(ctx, *tl.UserID, tl.Alias)
// 	require.NoError(t, err)

// 	_, err = repo.Get(ctx, tl.ID)
// 	require.ErrorIs(t, err, constants.ErrNotFound)
// }
