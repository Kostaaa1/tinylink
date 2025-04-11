package tinylink_test

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	"github.com/Kostaaa1/tinylink/internal/domain/user"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/db/sqlitedb"
	"github.com/Kostaaa1/tinylink/pkg/config"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

var (
	db *sqlx.DB
)

func init() {
	dbPath := "./tinylink_test.db"
	err := exec.Command("rm", "-rf", dbPath).Run()
	if err != nil {
		panic(err)
	}

	conf := config.SQLConfig{
		SQLitePath:   dbPath,
		MaxOpenConns: 25,
		MaxIdleConns: 25,
	}

	db, err = sqlitedb.StartDB(conf)
	if err != nil {
		panic(err)
	}

	file, err := os.ReadFile("../sql/tables.sql")
	if err != nil {
		panic(err)
	}
	db.Exec(string(file))
}

func TestTinylinkService(t *testing.T) {
	// ctx := context.Background()

	// userProvider := user.NewRepositoryProvider(db)
	// userAdapters := userProvider.GetDbAdapters()
	// userService := user.NewService(userProvider)
	// userDb := userAdapters.UserRepository

	// tlProvider := tinylink.NewRepositoryProvider(db, nil)
	// tlAdapters := tlProvider.GetAdapters()
	// tlService := tinylink.NewService(tlProvider)
	// tlDb := tlAdapters.TinylinkDBRepository

	// newUser := &user.User{
	// 	Email: "testuser@gmail.com",
	// 	Name:  "TestUser",
	// }
	// newUser.Password.Set("test123")
	// err := userDb.Insert(ctx, newUser)
	// require.Nil(t, err)

	// newTl := &tinylink.Tinylink{
	// 	Alias:       "service",
	// 	OriginalURL: "https://webmail.loopia.rs/?_task=mail&_mbox=INBOX.Spam",
	// 	UserID:      strconv.FormatUint(newUser.ID, 10),
	// }
	// err = tlDb.Insert(ctx, newTl)
	// require.Nil(t, err)

	// // for service layer
	// jwt, err := auth.GenerateJWT(newUser.ID, newUser.Email)
	// require.Nil(t, err)
	// claims, err := auth.VerifyJWT(jwt)
	// require.Nil(t, err)

	// ctx = context.WithValue(ctx, "claims", claims)

	// tl, err := tlService.Get(ctx, newTl.Alias)
	// require.Nil(t, err)

	// fmt.Println()
}

func TestTinylinKSQLite(t *testing.T) {
	ctx := context.Background()

	userProvider := user.NewRepositoryProvider(db)
	userAdapters := userProvider.GetAdapters()
	userDb := userAdapters.UserDbRepository

	provider := tinylink.NewRepositoryProvider(db, nil)
	adapters := provider.GetAdapters()
	tlDb := adapters.DBAdapters.TinylinkDBRepository

	// Inserting user
	newUser := &user.User{
		Email: "testuser@gmail.com",
		Name:  "TestUser",
	}
	newUser.Password.Set("test123")
	err := userDb.Insert(ctx, newUser)
	require.Nil(t, err)

	tl := &tinylink.Tinylink{
		UserID:      newUser.GetID(),
		OriginalURL: "https://codingchallenges.fyi/challenges/challenge-json-parser/",
		Alias:       "cc123",
		Private:     false,
	}
	err = tlDb.Insert(ctx, tl)
	require.Nil(t, err)
	require.NotZero(t, tl.CreatedAt)
	require.NotEmpty(t, tl.ID)

	tl, err = tlDb.Get(ctx, tl.Alias)
	require.Nil(t, err)
	require.NotEmpty(t, tl)

	// duplicate alias error
	err = tlDb.Insert(ctx, tl)
	require.NotNil(t, err)
	require.Equal(t, err, tinylink.ErrAliasExists)

	tl.Alias = "321cc"
	err = tlDb.Insert(ctx, tl)
	require.Nil(t, err)
	require.NotZero(t, tl.CreatedAt)
	require.NotEmpty(t, tl.ID)

	links, err := tlDb.List(ctx, newUser.GetID())
	require.Nil(t, err)
	require.Equal(t, len(links), 2)

	// UPDATE testing
	tl.Private = true
	tl.Alias = "kosta"
	tl.ExpiresAt = time.Now().Add(time.Hour * 1).Unix()
	prevVersion := tl.Version

	err = tlDb.Update(ctx, tl)
	require.Nil(t, err)
	require.True(t, tl.Private)
	require.Equal(t, tl.Version, prevVersion+1)
	require.NotEmpty(t, tl.ExpiresAt)

	fetched, err := tlDb.Get(ctx, tl.Alias)
	require.NotNil(t, err)
	require.Empty(t, fetched)

	newfetched, err := tlDb.GetByUserID(ctx, tl.UserID, tl.Alias)
	require.Nil(t, err)
	require.NotEmpty(t, newfetched)
	require.True(t, newfetched.Private)
	require.NotZero(t, newfetched.ExpiresAt)
	require.Equal(t, newfetched.Version, prevVersion+1)

	oldC := newfetched.UsageCount
	oldLV := newfetched.LastVisited

	time.Sleep(time.Second)

	// Usage count
	err = tlDb.UpdateUsage(ctx, newfetched)
	require.Nil(t, err)
	require.Equal(t, newfetched.UsageCount, oldC+1)
	require.NotEqual(t, newfetched.LastVisited, oldLV)
}
