package tinylink

import (
	"errors"
	"net/url"
	"time"

	"github.com/Kostaaa1/tinylink/internal/common/validator"
)

var (
	ErrAliasExists = errors.New("this alias is not available. All aliasses must be unique")
	ErrURLExists   = errors.New("you've already created a tinylink with this URL")
)

type Tinylink struct {
	ID          uint64 `json:"id"`
	Alias       string `json:"alias"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	Private     bool   `json:"private"`
	UsageCount  int    `json:"usage_count"`
	Domain      string `json:"domain"`
	Version     uint64 `json:"version"`
	LastVisited int64  `json:"last_visited"`
	ExpiresAt   int64  `json:"expires_at"`
	CreatedAt   int64  `json:"created_at"`
}

type TinylinkDTO struct {
	ID          uint64    `json:"id"`
	Alias       string    `json:"alias"`
	OriginalURL string    `json:"original_url"`
	UserID      string    `json:"user_id"`
	Private     bool      `json:"private"`
	UsageCount  int       `json:"usage_count"`
	Domain      string    `json:"domain"`
	Version     uint64    `json:"version"`
	LastVisited time.Time `json:"last_visited"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

func MapToTinylinkDTO(tl *Tinylink) TinylinkDTO {
	return TinylinkDTO{
		ID:          tl.ID,
		Alias:       tl.Alias,
		OriginalURL: tl.OriginalURL,
		UserID:      tl.UserID,
		Private:     tl.Private,
		UsageCount:  tl.UsageCount,
		Domain:      tl.Domain,
		Version:     tl.Version,
		CreatedAt:   time.Unix(tl.CreatedAt, 0),
		LastVisited: time.Unix(tl.LastVisited, 0),
		ExpiresAt:   time.Unix(tl.ExpiresAt, 0),
	}
}

type InsertTinylinkRequest struct {
	OriginalURL string `json:"url"`
	Alias       string `json:"alias"`
	Domain      string `json:"domain"`
	Private     bool   `json:"private"`
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

func (req *InsertTinylinkRequest) IsValid(v *validator.Validator) bool {
	v.Check(req.OriginalURL != "", "url", "must be provided")
	_, err := url.ParseRequestURI(req.OriginalURL)
	v.Check(err == nil, "url", "invalid url format")
	v.Check(!(req.Alias != "" && len(req.Alias) < 5), "alias", "must be at least 5 characters long")
	return v.Valid()
}
