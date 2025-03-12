package entities

import (
	"fmt"
	"net/http"
	"time"

	"github.com/skip2/go-qrcode"
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
	Tinylink    string    `json:"tinylink"`
	Alias       string    `json:"alias"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	QR          QR        `json:"qr"`
}

// add validation logic /maybe some helper function
func NewTinylink(domain, originalURL, alias string) (*Tinylink, error) {
	pngBytes, err := qrcode.Encode(fmt.Sprintf("%s/%s", domain, alias), qrcode.Medium, 127)
	if err != nil {
		fmt.Println("error while generating QR code", err)
		return nil, err
	}

	base64Bytes := []byte("data:image/png;base64,")
	base64Bytes = append(base64Bytes, pngBytes...)

	return &Tinylink{
		Tinylink:    fmt.Sprintf("%s/%s", domain, alias),
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
		Tinylink:    data["host"],
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
