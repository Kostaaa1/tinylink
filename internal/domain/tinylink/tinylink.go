package tinylink

import (
	"errors"
	"net/url"
	"time"

	"github.com/Kostaaa1/tinylink/pkg/validator"
)

var (
	anonTTL  = 30 * 24 * time.Hour
	cacheTTL = 7 * 24 * time.Hour

	ErrAliasExists = errors.New("this alias is not available. All aliasses must be unique")
	ErrURLExists   = errors.New("you've already created a tinylink with this URL")
)

type Tinylink struct {
	ID          uint64
	Alias       string
	OriginalURL string
	UserID      *string
	Private     bool
	UsageCount  int
	Domain      *string
	Version     uint64
	LastVisited int64
	ExpiresAt   int64
	CreatedAt   int64
}

type InsertTinylinkRequest struct {
	OriginalURL string `json:"url"`
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
