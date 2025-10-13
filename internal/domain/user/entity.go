package user

import (
	"errors"
	"regexp"
	"time"

	"github.com/Kostaaa1/tinylink/pkg/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	EmailRX               = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	ErrDuplicateEmail     = errors.New("duplicate email")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNoUserPasswordSet  = errors.New("password not set for user")
)

type GoogleUser struct {
	UserID        uint64
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	VerifiedEmail bool      `json:"verified_email"`
	Name          string    `json:"name"`
	GivenName     string    `json:"given_name"`
	FamilyName    string    `json:"family_name"`
	Picture       string    `json:"picture"`
	CreatedAt     time.Time `json:"created_at"`
}

type User struct {
	ID        uint64
	Name      string
	Email     string
	Password  password
	CreatedAt time.Time
	Version   int
	Google    *GoogleUser
}

func (u *User) SetPassword(pwHash []byte) {
	u.Password = password{Hash: pwHash}
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

func (u *User) HasPassword() bool { return len(u.Password.Hash) > 0 }

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(len(email) <= 254, "email", "must not be more then 254 bytes long")
	v.Check(validator.Matches(email, EmailRX), "email", "must be valid email address")
}

func ValidatePasswordPlainText(v *validator.Validator, plainText string) error {
	v.Check(plainText != "", "password", "must be provided")
	v.Check(len(plainText) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(plainText) <= 72, "password", "must not be more then 72 bytes long")
	return nil
}

func ValidatePasswordPlainTextWithKey(v *validator.Validator, key, plainText string) error {
	v.Check(plainText != "", key, "must be provided")
	v.Check(len(plainText) >= 8, key, "must be at least 8 bytes long")
	v.Check(len(plainText) <= 72, key, "must not be more then 72 bytes long")
	return nil
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must be provided")

	ValidateEmail(v, user.Email)
	if user.Password.plainText != nil {
		ValidatePasswordPlainText(v, *user.Password.plainText)
	}
	if user.Password.Hash == nil {
		panic("missing password hash for user")
	}
}
