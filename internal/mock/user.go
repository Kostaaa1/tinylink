package mock

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strconv"
	"testing"

	"github.com/Kostaaa1/tinylink/internal/domain/user"
	"github.com/stretchr/testify/require"
)

func InsertUser(t *testing.T, userRepo user.Repository) uint64 {
	userData := User()
	err := userRepo.Insert(context.Background(), userData)
	require.NoError(t, err)
	return userData.ID
}

func User() *user.User {
	val := rand.IntN(10000)

	userData := &user.User{
		ID:     uint64(val),
		Name:   fmt.Sprintf("testname_%d", val),
		Email:  fmt.Sprintf("testname_%d@gmail.com", val),
		Google: nil,
	}

	randPw := strconv.Itoa(rand.IntN(100000000))
	userData.SetPassword([]byte(randPw))
	return userData
}

func RandUserID() uint64 {
	return uint64(rand.IntN(10000))
}
