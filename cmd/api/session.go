package main

import (
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
)

type contextKey string

const tinylinkSessionKey contextKey = "tinylink_session"

func getSessionFromContext(r *http.Request) (*sessions.Session, error) {
	session, ok := r.Context().Value(tinylinkSessionKey).(*sessions.Session)
	if !ok {
		return nil, errors.New("no session found in context")
	}
	return session, nil
}

func getSessionID(r *http.Request) (string, error) {
	sess, err := getSessionFromContext(r)
	if err != nil {
		return "", err
	}
	s, ok := sess.Values["session_id"].(string)
	if ok {
		return s, nil
	}
	return "", errors.New("session is not a string?")
}
