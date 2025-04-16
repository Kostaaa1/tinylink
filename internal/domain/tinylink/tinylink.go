package tinylink

import (
	"errors"
	"net/url"
	"time"

	"github.com/Kostaaa1/tinylink/pkg/validator"
)

var (
	anonTTL  = 30 * 24 * time.Hour
	cacheTTL = 6 * time.Hour

	ErrAliasExists = errors.New("this alias is not available. All aliasses must be unique")
	ErrURLExists   = errors.New("you've already created a tinylink with this URL")
)

type Tinylink struct {
	ID          uint64
	Alias       string
	URL         string
	UserID      string
	Private     bool
	UsageCount  uint64
	Domain      *string
	Version     uint64
	LastVisited int64
	ExpiresAt   int64
	CreatedAt   int64
}

type InsertTinylinkRequest struct {
	URL     string `json:"url"`
	Alias   string `json:"alias"`
	Domain  string `json:"domain"`
	Private bool   `json:"private"`
}

func (req *InsertTinylinkRequest) IsValid(v *validator.Validator) bool {
	v.Check(req.URL != "", "url", "must be provided")
	_, err := url.ParseRequestURI(req.URL)
	v.Check(err == nil, "url", "invalid url format")
	v.Check(!(req.Alias != "" && len(req.Alias) < 5), "alias", "must be at least 5 characters long")
	return v.Valid()
}

type UpdateTinylinkRequest struct {
	ID      uint64  `json:"id"`
	URL     *string `json:"url"`
	Alias   *string `json:"alias"`
	Domain  *string `json:"domain"`
	Private bool    `json:"private"`
}

func (req *UpdateTinylinkRequest) IsValid(v *validator.Validator) bool {
	v.Check(req.ID != 0, "id", "must be provided")
	if req.Alias != nil {
		v.Check(*req.Alias != "", "alias", "must be provided")
		v.Check(!(*req.Alias != "" && len(*req.Alias) < 5), "alias", "must be at least 5 characters long")
	}
	if req.URL != nil {
		_, err := url.Parse(*req.URL)
		v.Check(err == nil, "url", "wrong URL format")
	}
	if req.Domain != nil {
		_, err := url.Parse(*req.Domain)
		v.Check(err == nil, "domain", "wrong URL format")
	}
	return v.Valid()
}
