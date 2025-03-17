package services

import (
	"github.com/Kostaaa1/tinylink/db"
	"github.com/Kostaaa1/tinylink/internal/data"
)

type UserService struct {
	User db.UserStore
}

func NewUserService(userStore db.UserStore) *UserService {
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
