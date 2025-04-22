package tinylink

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/Kostaaa1/tinylink/pkg/validator"
)

var (
	cacheTTL       = 6 * time.Hour
	ErrAliasExists = errors.New("this alias is not available. All aliasses must be unique")
	ErrURLExists   = errors.New("you've already created a tinylink with this URL")
)

type Tinylink struct {
	ID        uint64
	Alias     string
	URL       string
	UserID    string
	Domain    *string
	Private   bool
	Version   uint64
	ExpiresAt int64
	CreatedAt int64
}

func ToMap(tl *Tinylink) map[string]interface{} {
	return map[string]interface{}{
		"id":         tl.ID,
		"url":        tl.URL,
		"alias":      tl.Alias,
		"private":    tl.Private,
		"domain":     tl.Domain,
		"version":    tl.Version,
		"expires_at": tl.ExpiresAt,
		"created_at": tl.CreatedAt,
	}
}

func FromMap(data map[string]string) (*Tinylink, error) {
	tl := &Tinylink{}

	if id, ok := data["id"]; ok {
		parsedID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			return nil, errors.New("invalid id format")
		}
		tl.ID = parsedID
	}

	if url, ok := data["url"]; ok {
		tl.URL = url
	}

	if alias, ok := data["alias"]; ok {
		tl.Alias = alias
	}

	if private, ok := data["private"]; ok {
		parsedPrivate, err := strconv.ParseBool(private)
		if err != nil {
			return nil, errors.New("invalid private format")
		}
		tl.Private = parsedPrivate
	}

	if domain, ok := data["domain"]; ok {
		tl.Domain = &domain
	}

	if version, ok := data["version"]; ok {
		parsedVersion, err := strconv.ParseUint(version, 10, 64)
		if err != nil {
			return nil, errors.New("invalid version format")
		}
		tl.Version = parsedVersion
	}

	if expiresAt, ok := data["expires_at"]; ok {
		parsedExpiresAt, err := strconv.ParseInt(expiresAt, 10, 64)
		if err != nil {
			return nil, errors.New("invalid expires_at format")
		}
		tl.ExpiresAt = parsedExpiresAt
	}

	if createdAt, ok := data["created_at"]; ok {
		parsedCreatedAt, err := strconv.ParseInt(createdAt, 10, 64)
		if err != nil {
			return nil, errors.New("invalid created_at format")
		}
		tl.CreatedAt = parsedCreatedAt
	}

	return tl, nil
}

type CreateTinylinkRequest struct {
	URL     string  `json:"url"`
	Alias   string  `json:"alias"`
	Private bool    `json:"private"`
	Domain  *string `json:"domain"`
}

func (req *CreateTinylinkRequest) IsValid(v *validator.Validator) bool {
	v.Check(req.URL != "", "url", "must be provided")
	_, err := url.ParseRequestURI(req.URL)
	v.Check(err == nil, "url", "invalid url format")
	v.Check(!(req.Alias != "" && len(req.Alias) < 5), "alias", "must be at least 5 characters long")
	return v.Valid()
}

type UpdateTinylinkRequest struct {
	URL     *string `json:"url"`
	Alias   *string `json:"alias"`
	Domain  *string `json:"domain"`
	Private bool    `json:"private"`
}

func (req *UpdateTinylinkRequest) IsValid(v *validator.Validator) bool {
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
