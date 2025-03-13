package data

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Kostaaa1/tinylink/internal/validator"
	"github.com/skip2/go-qrcode"
)

var (
	ErrAliasExists = errors.New("this alias is not available. All aliasses must be unique")
	ErrURLExists   = errors.New("you've already created a tinylink with this URL")
)

type QueryParams struct {
	SessionID string
	Alias     string
}

type QR struct {
	// Data     []byte `json:"data"`
	Base64   []byte `json:"base64"`
	Width    string `json:"width"`
	Height   string `json:"height"`
	Size     string `json:"size"`
	MimeType string `json:"mimetype"`
}

type Tinylink struct {
	// Tinylink    string    `json:"tinylink"`
	Alias       string    `json:"alias"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	QR          QR        `json:"qr"`
}

// add validation logic for tinylink /maybe some helper function

func NewTinylink(domain, originalURL, alias string) (*Tinylink, error) {
	pngBytes, err := qrcode.Encode(fmt.Sprintf("%s/%s", domain, alias), qrcode.Medium, 127)
	if err != nil {
		return nil, err
	}

	base64Bytes := []byte("data:image/png;base64,")
	base64Bytes = append(base64Bytes, pngBytes...)

	return &Tinylink{
		// Tinylink:    fmt.Sprintf("%s/%s", domain, alias),
		Alias:       alias,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
		QR: QR{
			Base64:   base64Bytes,
			Width:    "127",
			Height:   "127",
			Size:     fmt.Sprintf("%d bytes", len(pngBytes)),
			MimeType: http.DetectContentType(pngBytes),
		},
	}, nil
}

func MapToTinylink(data map[string]string) *Tinylink {
	return &Tinylink{
		// Tinylink:    data["host"],
		Alias:       data["alias"],
		OriginalURL: data["original_url"],
		QR: QR{
			Base64:   []byte(data["qr:data"]),
			Width:    data["qr:width"],
			Height:   data["qr:height"],
			Size:     data["qr:size"],
			MimeType: data["qr:mimetype"],
		},
	}
}

type CreateTinylinkRequest struct {
	URL    string `json:"url"`
	Alias  string `json:"alias"`
	Domain string `json:"domain"`
}

func (req *CreateTinylinkRequest) IsValid(v *validator.Validator) bool {
	v.Check(req.URL != "", "url", "must be provided")
	_, err := url.ParseRequestURI(req.URL)
	v.Check(err == nil, "url", "invalid url format")
	v.Check(!(req.Alias != "" && len(req.Alias) < 5), "alias", "must be at least 5 characters long")
	return v.Valid()
}
