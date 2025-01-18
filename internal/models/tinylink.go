package models

type QR struct {
	Data     []byte `json:"data"`
	Width    string `json:"width"`
	Height   string `json:"height"`
	Size     string `json:"size"`
	MimeType string `json:"mimetype"`
}

type Tinylink struct {
	Tinylink    string `json:"tinylink"`
	Alias       string `json:"alias"`
	OriginalURL string `json:"original_url"`
	QR          QR     `json:"qr"`
}
