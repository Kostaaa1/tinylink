package handlers

import (
	"net"
	"net/http"
)

// func createHashAlias(clientID, url string, length int) string {
// 	s := clientID + url
// 	return fmt.Sprintf("%x", sha1.Sum([]byte(s)))[:length]
// }

func getServerURL() string {
	return "http://localhost:3000"
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func getUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

func getReferer(r *http.Request) string {
	return r.Referer()
}
