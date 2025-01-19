package tinylink

import (
	"time"
)

type QR struct {
	Data     []byte `json:"data"`
	Width    string `json:"width"`
	Height   string `json:"height"`
	Size     string `json:"size"`
	MimeType string `json:"mimetype"`
}

type Tinylink struct {
	Tinylink    string    `json:"tinylink"`
	Alias       string    `json:"alias"`
	OriginalURL string    `json:"original_url"`
	QR          QR        `json:"qr"`
	CreatedAt   time.Time `json:"created_at"`
}

type QueryParams struct {
	ClientID string
	Alias    string
}

type TinylinkRepository interface {
	GetAll(qp QueryParams) ([]*Tinylink, error)
	Delete(qp QueryParams) error
	Create(tl *Tinylink, qp QueryParams) error
	Get(qp QueryParams) (*Tinylink, error)
	ValidateAlias(alias string) error
	ValidateOriginalURL(clientID, URL string) error
	////////////////
	// Ping(ctx context.Context) error
}
