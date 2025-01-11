package data

type QR struct {
	ImageURL string `json:"image_url"`
	Width    string `json:"width"`
	Height   string `json:"height"`
}

type Tinylink struct {
	URL    string `json:"url"`
	Domain string `json:"domain"`
	Alias  string `json:"alias"`
	QR     QR     `json:"qr"`
}
