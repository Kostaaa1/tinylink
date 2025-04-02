package tests

import (
	"context"
	"testing"
	"time"

	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/Kostaaa1/tinylink/internal/domain/user"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/db/sqlitedb"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestNormalAuthFlow(t *testing.T) {

}

func TestUserDBAuthFlow(t *testing.T) {
	t.Parallel()

	input := struct {
		name     string
		email    string
		password string
	}{
		name:     "Kosta",
		email:    "kostaarsic123@gmail.com",
		password: "Kosta123",
	}

	newUser := &user.User{
		Name:      input.name,
		Email:     input.email,
		CreatedAt: time.Now(),
	}
	err := newUser.Password.Set(input.password)
	require.Nil(t, err)

	db, err := sqlx.Connect("sqlite3", "../tinylink.db")
	require.Nil(t, err)

	store := sqlitedb.NewRepositoriesFromDB(db)

	ctx := context.Background()

	userByEmail, err := store.User.GetByEmail(ctx, newUser.Email)
	require.Equal(t, err, data.ErrRecordNotFound)
	require.Nil(t, userByEmail)

	err = store.User.Insert(ctx, newUser)
	require.Nil(t, err)

	userByEmail, err = store.User.GetByEmail(ctx, newUser.Email)
	require.Nil(t, err)

	require.NotZero(t, userByEmail.ID, "ID should not be empty")
	require.NotEmpty(t, userByEmail.Name, "Name should not be empty")
	require.NotEmpty(t, userByEmail.Email, "Email should not be empty")
	require.NotEmpty(t, userByEmail.Password.Hash, "Password hash should not be empty")
	require.Empty(t, userByEmail.Google, "Google struct should be empty")
	require.False(t, newUser.CreatedAt.IsZero(), "CreatedAt should not be zero time")

	newUser.Name = "KostaTest"
	currentVersion := newUser.Version
	err = store.User.Update(ctx, newUser)
	require.Nil(t, err)
	require.Equal(t, currentVersion+1, newUser.Version)

	// when userdata gets updated, version should be increased

	// googleUser := &user.GoogleUser{
	// 	UserID:        1,
	// 	ID:            "3213213",
	// 	Email:         "kostaarsic123@gmail.com",
	// 	Name:          "Kosta",
	// 	GivenName:     "Kosta",
	// 	FamilyName:    "Arsic",
	// 	Picture:       "https://test.com",
	// 	CreatedAt:     time.Now(),
	// 	VerifiedEmail: true,
	// }
}
