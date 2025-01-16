package models

type QR struct {
	Data     []byte `json:"data"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Size     int    `json:"size"`
	MimeType string `json:"mimetype"`
}

type Tinylink struct {
	Host        string `json:"host"`
	Alias       string `json:"alias"`
	OriginalURL string `json:"original_url"`
	QR          QR     `json:"qr"`
}
