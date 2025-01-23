package main

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/gorilla/sessions"
)

type envelope map[string]interface{}

func (a *app) writeJSON(w http.ResponseWriter, status int, data interface{}, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	js = append(js, '\n')
	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (a *app) readJSON(r *http.Request, dst interface{}) error {
	err := json.NewDecoder(r.Body).Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}
	return nil
}

type contextKey string

const tinylinkSessionKey contextKey = "tinylink_session"

func getSessionID(r *http.Request) (string, error) {
	session, ok := r.Context().Value(tinylinkSessionKey).(*sessions.Session)
	if !ok {
		return "", errors.New("no session found in context")
	}
	s, ok := session.Values["session_id"].(string)
	if ok {
		return s, nil
	}
	return "", errors.New("session is not a string?")
}

func generateRandHex(l int) string {
	b := make([]byte, l)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func createHashAlias(clientID, url string, length int) string {
	s := clientID + url
	return fmt.Sprintf("%x", sha1.Sum([]byte(s)))[:length]
}

func (a *app) getServerURL() string {
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
