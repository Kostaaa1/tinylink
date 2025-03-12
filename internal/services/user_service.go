package services

import (
	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/Kostaaa1/tinylink/internal/store"
)

type UserService struct {
	UserRepo store.UserRepository
}

func NewUserService(userRepo store.UserRepository) *UserService {
	return &UserService{
		UserRepo: userRepo,
	}
}

func (s *UserService) Register(user *data.User) error {
	return s.UserRepo.Insert(user)
}
