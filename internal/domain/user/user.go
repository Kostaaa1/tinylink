package user

import (
	"errors"
	"strconv"
	"time"

	"github.com/Kostaaa1/tinylink/pkg/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail     = errors.New("duplicate email")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNoUserPasswordSet  = errors.New("password not set for user")
)

func NewUserDTO(user *User) UserDTO {
	dto := UserDTO{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: time.Unix(user.CreatedAt, 0),
	}
	if user.Google != nil {
		dto.Google = &GoogleUserDTO{
			ID:            user.Google.ID,
			Name:          user.Google.Name,
			Picture:       user.Google.Picture,
			FamilyName:    user.Google.FamilyName,
			GivenName:     user.Google.GivenName,
			VerifiedEmail: user.Google.VerifiedEmail,
			CreatedAt:     time.Unix(user.Google.CreatedAt, 0),
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

type GoogleUser struct {
	UserID        uint64 `json:"-"`
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"is_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	CreatedAt     int64  `json:"created_at"`
}

type User struct {
	ID        uint64      `json:"id"`
	Name      string      `json:"name"`
	Email     string      `json:"email"`
	Password  password    `json:"-"`
	CreatedAt int64       `json:"created_at"`
	Version   int         `json:"version"`
	Google    *GoogleUser `json:"google"`
}

type password struct {
	plainText *string
	Hash      []byte
}

func (p *password) Set(plainPW string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPW), 12)
	if err != nil {
		return err
	}

	p.plainText = &plainPW
	p.Hash = hash

	return nil
}

func (p *password) Matches(plainPW string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hash, []byte(plainPW))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, err
		default:
			return false, err
		}
	}
	return true, nil
}

func (u *User) GetID() string {
	if u == nil {
		return ""
	}
	return strconv.FormatUint(u.ID, 10)
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(len(email) <= 254, "email", "must not be more then 254 bytes long")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be valid email address")
}

func ValidatePasswordPlainText(v *validator.Validator, plainText string) {
	v.Check(plainText != "", "password", "must be provided")
	v.Check(len(plainText) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(plainText) <= 72, "password", "must not be more then 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more then 500 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password.plainText != nil {
		ValidatePasswordPlainText(v, *user.Password.plainText)
	}
	if user.Password.Hash == nil {
		panic("missing password hash for user")
	}
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
