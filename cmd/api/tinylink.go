package main

import (
	"fmt"
	"net/http"

	"github.com/Kostaaa1/tinylink/internal/models"
	"github.com/skip2/go-qrcode"
)

func (a *app) newTinylink(URL, alias string) (*models.Tinylink, error) {
	pngBytes, err := qrcode.Encode(fmt.Sprintf("http://localhost:%s/%s", a.config.port, alias), qrcode.Medium, 127)
	if err != nil {
		return nil, err
	}

	return &models.Tinylink{
		Host:        "http://localhost:3000",
		Alias:       alias,
		OriginalURL: URL,
		QR: models.QR{
			Data:     pngBytes,
			Width:    127,
			Height:   127,
			Size:     len(pngBytes),
			MimeType: http.DetectContentType(pngBytes),
		},
	}, nil
}
