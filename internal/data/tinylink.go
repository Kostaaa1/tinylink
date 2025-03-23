package data

import (
	"crypto/sha1"
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
	Data     []byte `db:"data" json:"data"`
	Width    string `db:"width" json:"width"`
	Height   string `db:"height" json:"height"`
	Size     string `db:"size" json:"size"`
	MimeType string `db:"mimetype" json:"mimetype"`
}

type Tinylink struct {
	ID        string    `db:"id" json:"id"`
	Alias     string    `db:"alias" json:"alias"`
	URL       string    `db:"original_url" json:"original_url"`
	UserID    string    `db:"user_id" json:"user_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	QR        QR        `db:"qr" json:"qr"`
}

// not effective?
func generateAlias(userID, URL string) string {
	s := fmt.Sprintf("%s%s%d", userID, URL, time.Now().Nanosecond())
	return fmt.Sprintf("%x", sha1.Sum([]byte(s)))[:8]
}

// add validation logic for tinylink /maybe some helper function
func NewTinylink(userID, domain, targetURL, alias string) (*Tinylink, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	if alias == "" {
		alias = generateAlias(userID, targetURL)
	}

	u := fmt.Sprintf("%s/%s", domain, alias)
	pngBytes, err := qrcode.Encode(u, qrcode.Medium, 127)
	if err != nil {
		return nil, err
	}
	base64Bytes := []byte("data:image/png;base64,")
	base64Bytes = append(base64Bytes, pngBytes...)

	return &Tinylink{
		Alias:  alias,
		URL:    parsedURL.String(),
		UserID: userID,
		QR: QR{
			Width:    "127",
			Height:   "127",
			Data:     base64Bytes,
			Size:     fmt.Sprintf("%d bytes", len(pngBytes)),
			MimeType: http.DetectContentType(pngBytes),
		},
		CreatedAt: time.Now(),
	}, nil
}

func MapToTinylink(data map[string]string) (*Tinylink, error) {
	url, err := url.Parse(data["url"])
	if err != nil {
		return nil, err
	}
	return &Tinylink{
		Alias: data["alias"],
		URL:   url.String(),
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
