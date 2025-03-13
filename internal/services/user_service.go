package services

import (
	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/Kostaaa1/tinylink/internal/store"
)

type UserService struct {
	User store.UserStore
}

func NewUserService(userStore store.UserStore) *UserService {
	return &UserService{
		User: userStore,
	}
}

func (s *UserService) GetByEmail(email string) (*data.User, error) {
	return s.User.GetByEmail(email)
}

func (s *UserService) Register(user *data.User) error {
	return s.User.Insert(user)
}
