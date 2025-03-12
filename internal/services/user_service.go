package services

import (
	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/Kostaaa1/tinylink/internal/store"
)

type UserService struct {
	UserStore store.UserStore
}

func NewUserService(userStore store.UserStore) *UserService {
	return &UserService{
		UserStore: userStore,
	}
}

func (s *UserService) Register(user *data.User) error {
	return s.UserStore.Insert(user)
}
