package api

import (
	"time"

	"github.com/Kostaaa1/tinylink/internal/domain/user"
)

type UserRegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserLoginRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func UserResponse(user *user.User) UserDTO {
	dto := UserDTO{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	}

	if user.Google != nil {
		dto.Google = &GoogleUserDTO{
			ID:            user.Google.ID,
			Name:          user.Google.Name,
			Picture:       user.Google.Picture,
			FamilyName:    user.Google.FamilyName,
			GivenName:     user.Google.GivenName,
			VerifiedEmail: user.Google.VerifiedEmail,
			CreatedAt:     user.Google.CreatedAt,
		}
	}

	return dto
}

type UserDTO struct {
	ID        uint64         `json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
	Google    *GoogleUserDTO `json:"google,omitempty"`
}

type GoogleUserDTO struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	GivenName     string    `json:"given_name"`
	FamilyName    string    `json:"family_name"`
	Picture       string    `json:"picture"`
	VerifiedEmail bool      `json:"is_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

type Role int

const (
	Admin Role = iota
	Subscriber
	User
	Guest
)

// type Permission int
// const (
// 	Read Permission = iota
// 	Create
// 	Update
// 	Delete
// 	BulkInsert
// 	Metrics
// )
