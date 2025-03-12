package data

import (
	"errors"
	"time"

	"github.com/Kostaaa1/tinylink/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrRecordExists   = errors.New("record not found")
	ErrDuplicateEmail = errors.New("duplicate email")
)

type User struct {
	ID        uint64    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

type password struct {
	plainText *string
	Hash      []byte
}

func (p *password) Set(plainText string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainText), 12)
	if err != nil {
		return err
	}

	p.plainText = &plainText
	p.Hash = hash

	return nil
}

func (p *password) Matches(plainTextPW string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hash, []byte(plainTextPW))
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
		panic("missing password has for user")
	}
}
