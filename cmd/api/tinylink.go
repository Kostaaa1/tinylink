package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Kostaaa1/tinylink/internal/domain"
	"github.com/skip2/go-qrcode"
)

func (a *app) newTinylink(URL, alias string) (*domain.Tinylink, error) {
	pngBytes, err := qrcode.Encode(fmt.Sprintf("http://localhost:%s/%s", a.cfg.port, alias), qrcode.Medium, 127)
	if err != nil {
		return nil, err
	}
	return &domain.Tinylink{
		Tinylink:    fmt.Sprintf("http://localhost:%s/%s", a.cfg.port, alias), // use *app
		Alias:       alias,
		OriginalURL: URL,
		QR: domain.QR{
			Data:     pngBytes,
			Width:    "127",
			Height:   "127",
			Size:     fmt.Sprintf("%d bytes", len(pngBytes)),
			MimeType: http.DetectContentType(pngBytes),
		},
		CreatedAt: time.Now(),
	}, nil
}
