package tinylink_test

import (
	"context"
	"os"
	"os/exec"
	"testing"

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
	err := exec.Command("rm", "-rf", "./tinylink.db").Run()
	if err != nil {
		panic(err)
	}

	conf := config.SQLConfig{
		SQLitePath:   "./tinylink.db",
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
	_, err = db.Exec(string(file))
}

func TestTinylinKSQLite(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	userProvider := user.NewRepositoryProvider(db)
	userAdapters := userProvider.GetDbAdapters()
	userDb := userAdapters.UserRepository

	newUser := &user.User{
		Email: "kosta.arsic123@gmail.com",
		Name:  "Kosta",
	}
	newUser.Password.Set("test123")
	err := userDb.Insert(ctx, newUser)
	require.Nil(t, err)

	provider := tinylink.NewRepositoryProvider(db, nil)
	adapters := provider.GetAdapters()
	tlDb := adapters.DBAdapters.TinylinkDBRepository

	tl := &tinylink.Tinylink{
		UserID:      newUser.GetID(),
		OriginalURL: "https://codingchallenges.fyi/challenges/challenge-json-parser/",
		Alias:       "cc123",
		Private:     false,
	}

	err = tlDb.Insert(ctx, tl)
	require.Nil(t, err)

	tl, err = tlDb.Get(ctx, tl.Alias)
	require.Nil(t, err)
	require.NotEmpty(t, tl)
	// test other fields

	tl = &tinylink.Tinylink{
		UserID:      newUser.GetID(),
		Alias:       "123cc",
		OriginalURL: "https://codingchallenges.fyi/challenges/challenge-json-parser/",
		Private:     false,
	}
	err = tlDb.Insert(ctx, tl)
	require.Nil(t, err)

	links, err := tlDb.List(ctx, newUser.GetID())
	require.Nil(t, err)
	require.Equal(t, len(links), 2)
}
