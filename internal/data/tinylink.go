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

type QR struct {
	Data     []byte `json:"data"`
	Width    string `json:"width"`
	Height   string `json:"height"`
	Size     string `json:"size"`
	MimeType string `json:"mimetype"`
}

type Tinylink struct {
	Alias     string    `json:"alias"`
	URL       *url.URL  `json:"original_url"`
	CreatedAt time.Time `json:"created_at"`
	QR        QR        `json:"qr"`
}

// add validation logic for tinylink /maybe some helper function

func NewTinylink(domain, originalURL, alias string) (*Tinylink, error) {
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return nil, err
	}

	pngBytes, err := qrcode.Encode(fmt.Sprintf("%s/%s", domain, alias), qrcode.Medium, 127)
	if err != nil {
		return nil, err
	}

	base64Bytes := []byte("data:image/png;base64,")
	base64Bytes = append(base64Bytes, pngBytes...)

	return &Tinylink{
		Alias:     alias,
		URL:       parsedURL,
		CreatedAt: time.Now(),
		QR: QR{
			Data:     base64Bytes,
			Width:    "127",
			Height:   "127",
			Size:     fmt.Sprintf("%d bytes", len(pngBytes)),
			MimeType: http.DetectContentType(pngBytes),
		},
	}, nil
}

func MapToTinylink(data map[string]string) (*Tinylink, error) {
	url, err := url.Parse(data["url"])
	if err != nil {
		return nil, err
	}
	return &Tinylink{
		Alias: data["alias"],
		URL:   url,
		QR: QR{
			Data:     []byte(data["qr:data"]),
			Width:    data["qr:width"],
			Height:   data["qr:height"],
			Size:     data["qr:size"],
			MimeType: data["qr:mimetype"],
		},
	}, nil
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
