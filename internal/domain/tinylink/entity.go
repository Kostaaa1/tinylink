package tinylink

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrAliasNotProvided = errors.New("alias not provided")
	ErrAliasExists      = errors.New("alias already exists")
	defaultTTL          = time.Hour
)

type Tinylink struct {
	ID         uint64     `json:"id"`
	Alias      string     `json:"alias"`
	URL        string     `json:"url"`
	UserID     *uint64    `json:"user_id"`
	GuestUUID  string     `json:"guest_id"`
	Domain     string     `json:"domain,omitempty"`
	Private    bool       `json:"private"`
	Version    uint64     `json:"version"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
	Expiration *time.Time `json:"expiration,omitempty"`
}

func (tl *Tinylink) ToMap() map[string]string {
	m := map[string]string{
		"id":         fmt.Sprintf("%d", tl.ID),
		"alias":      tl.Alias,
		"url":        tl.URL,
		"user_id":    fmt.Sprintf("%d", tl.UserID),
		"domain":     tl.Domain,
		"private":    fmt.Sprintf("%t", tl.Private),
		"version":    fmt.Sprintf("%d", tl.Version),
		"created_at": tl.CreatedAt.Format(time.RFC3339),
		// "created_at": strconv.Itoa(int(tl.CreatedAt.Unix())),
	}
	if tl.UpdatedAt != nil {
		m["updated_at"] = tl.UpdatedAt.Format(time.RFC3339)
	}
	if tl.Expiration != nil {
		m["expiration"] = tl.Expiration.Format(time.RFC3339)
	}
	return m
}

func FromMap(m map[string]string) (*Tinylink, error) {
	var err error
	var tl Tinylink

	if id, ok := m["id"]; ok {
		_, err = fmt.Sscanf(id, "%d", &tl.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid id: %w", err)
		}
	}

	if alias, ok := m["alias"]; ok {
		tl.Alias = alias
	}

	if url, ok := m["url"]; ok {
		tl.URL = url
	}

	if userID, ok := m["user_id"]; ok {
		_, err = fmt.Sscanf(userID, "%d", &tl.UserID)
		if err != nil {
			return nil, fmt.Errorf("invalid user_id: %w", err)
		}
	}

	if domain, ok := m["domain"]; ok {
		tl.Domain = domain
	}

	if priv, ok := m["private"]; ok {
		tl.Private = priv == "true"
	}

	if version, ok := m["version"]; ok {
		_, err = fmt.Sscanf(version, "%d", &tl.Version)
		if err != nil {
			return nil, fmt.Errorf("invalid version: %w", err)
		}
	}

	if createdAt, ok := m["created_at"]; ok {
		tl.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("invalid created_at: %w", err)
		}
	}

	if updatedAt, ok := m["updated_at"]; ok {
		t := new(time.Time)
		*t, err = time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid updated_at: %w", err)
		}
		tl.UpdatedAt = t
	}

	if expiration, ok := m["expiration"]; ok {
		t := new(time.Time)
		*t, err = time.Parse(time.RFC3339, expiration)
		if err != nil {
			return nil, fmt.Errorf("invalid expiration: %w", err)
		}
		tl.Expiration = t
	}

	return &tl, nil
}

type RedirectValue struct {
	RowID uint64
	Alias string
	URL   string
}
