package user

import (
	"errors"
	"strconv"

	"github.com/Kostaaa1/tinylink/pkg/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail     = errors.New("duplicate email")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNoUserPasswordSet  = errors.New("password not set for user")
)

type GoogleUser struct {
	UserID        uint64
	ID            string
	Email         string
	VerifiedEmail bool
	Name          string
	GivenName     string
	FamilyName    string
	Picture       string
	CreatedAt     int64
}

type User struct {
	ID        uint64
	Name      string
	Email     string
	Password  password
	CreatedAt int64
	Version   int
	Google    *GoogleUser
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
