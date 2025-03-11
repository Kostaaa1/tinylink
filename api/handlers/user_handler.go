package handlers

import (
	"github.com/Kostaaa1/tinylink/internal/services"
)

type UserHandler struct {
	service *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		service: userService,
	}
}
