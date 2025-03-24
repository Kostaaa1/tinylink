package data

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/Kostaaa1/tinylink/internal/validator"
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
	ID         uint64    `json:"id"`
	Alias      string    `json:"alias"`
	URL        string    `json:"original_url"`
	UserID     string    `json:"user_id"`
	Public     bool      `json:"public"`
	CreatedAt  time.Time `json:"created_at"`
	UsageCount int       `json:"usage_count,omitempty"`
	Domain     string    `json:"domain,omitempty"`
	QR         *QR       `json:"qr,omitempty"`
}

func GenerateAlias(userID, URL string) string {
	s := fmt.Sprintf("%s%s%d", userID, URL, time.Now().Nanosecond())
	return fmt.Sprintf("%x", sha1.Sum([]byte(s)))[:8]
}

func MapToTinylink(data map[string]string) (*Tinylink, error) {
	url, err := url.Parse(data["url"])
	if err != nil {
		return nil, err
	}
	return &Tinylink{
		Alias: data["alias"],
		URL:   url.String(),
		QR: &QR{
			Data:     []byte(data["qr:data"]),
			Width:    data["qr:width"],
			Height:   data["qr:height"],
			Size:     data["qr:size"],
			MimeType: data["qr:mimetype"],
		},
	}, nil
}

type InsertTinylinkRequest struct {
	URL    string `json:"url"`
	Alias  string `json:"alias"`
	Domain string `json:"domain"`
	Public bool   `json:"public"`
}

func (req *InsertTinylinkRequest) IsValid(v *validator.Validator) bool {
	v.Check(req.URL != "", "url", "must be provided")
	_, err := url.ParseRequestURI(req.URL)
	v.Check(err == nil, "url", "invalid url format")
	v.Check(!(req.Alias != "" && len(req.Alias) < 5), "alias", "must be at least 5 characters long")
	return v.Valid()
}

type UpdateTinylinkRequest struct {
	ID     uint64 `json:"id"`
	Alias  string `json:"alias"`
	Public bool   `json:"public"`
	Domain string `json:"domain"`
}

func (req *UpdateTinylinkRequest) IsValid(v *validator.Validator) bool {
	v.Check(req.ID != 0, "id", "must be provided")
	v.Check(req.Alias != "", "alias", "must be provided")
	v.Check(!(req.Alias != "" && len(req.Alias) < 5), "alias", "must be at least 5 characters long")
	return v.Valid()
}
