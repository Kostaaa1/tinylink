package data

type QR struct {
	ImageURL string `json:"image_url"`
	Width    string `json:"width"`
	Height   string `json:"height"`
}

type Tinylink struct {
	ID      string `json:"id"`
	TinyURL string `json:"tiny_url"`
	URL     string `json:"url"`
	QR      QR     `json:"qr"`
}
