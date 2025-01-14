package main

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/skip2/go-qrcode"
)

func main() {
	u := "https://twitch.tv/emiru"
	b, err := qrcode.Encode(u, qrcode.High, 256)
	if err != nil {
		panic(err)
	}
	base64Image := base64.StdEncoding.EncodeToString(b)
	fmt.Println(base64Image)
	mimeType := http.DetectContentType(b)
	fmt.Println(mimeType)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image)
	fmt.Println(dataURL)
}
