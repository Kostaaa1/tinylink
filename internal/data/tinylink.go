package data

import (
	"errors"
	"net/url"
	"time"

	"github.com/Kostaaa1/tinylink/internal/validator"
)

var (
	ErrAliasExists = errors.New("this alias is not available. All aliasses must be unique")
	ErrURLExists   = errors.New("you've already created a tinylink with this URL")
)

type Tinylink struct {
	ID          uint64    `json:"id"`
	Alias       string    `json:"alias"`
	OriginalURL string    `json:"original_url"`
	UserID      string    `json:"user_id"`
	Private     bool      `json:"private"`
	LastVisited time.Time `json:"last_visited"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	UsageCount  int       `json:"usage_count,omitempty"`
	Domain      string    `json:"domain,omitempty"`
}

func MapToTinylink(data map[string]string) (*Tinylink, error) {
	url, err := url.Parse(data["url"])
	if err != nil {
		return nil, err
	}
	return &Tinylink{
		Alias:       data["alias"],
		OriginalURL: url.String(),
	}, nil
}

type InsertTinylinkRequest struct {
	OriginalURL string `json:"original_url"`
	Alias       string `json:"alias"`
	Domain      string `json:"domain"`
	Private     bool   `json:"private"`
}

func (req *InsertTinylinkRequest) IsValid(v *validator.Validator) bool {
	v.Check(req.OriginalURL != "", "url", "must be provided")
	_, err := url.ParseRequestURI(req.OriginalURL)
	v.Check(err == nil, "url", "invalid url format")
	v.Check(!(req.Alias != "" && len(req.Alias) < 5), "alias", "must be at least 5 characters long")
	return v.Valid()
}

type UpdateTinylinkRequest struct {
	ID      uint64 `json:"id"`
	Alias   string `json:"alias"`
	Private bool   `json:"private"`
	Domain  string `json:"domain"`
}

func (req *UpdateTinylinkRequest) IsValid(v *validator.Validator) bool {
	v.Check(req.ID != 0, "id", "must be provided")
	v.Check(req.Alias != "", "alias", "must be provided")
	v.Check(!(req.Alias != "" && len(req.Alias) < 5), "alias", "must be at least 5 characters long")
	return v.Valid()
}
