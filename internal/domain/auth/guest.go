package auth

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

var (
	guestCookieName = "guest_uuid"
	MaxAge          = 60 * 60 * 24 * 7
	GuestIDDuration = time.Duration(MaxAge) * time.Second
)

func GetGuestUUID(r *http.Request) *string {
	cookie, err := r.Cookie(guestCookieName)
	if err == nil {
		return &cookie.Value
	}
	return nil
}

func EnsureGuestUUID(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie(guestCookieName)
	if err == nil {
		return cookie.Value
	}

	guestID := uuid.NewString()
	http.SetCookie(w, &http.Cookie{
		Name:     guestCookieName,
		Value:    guestID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   MaxAge,
	})

	return guestID
}

type UserSig struct {
	GuestUUID *string
	UserID    *uint64
}

func GetUserSignature(r *http.Request) UserSig {
	sig := UserSig{GuestUUID: GetGuestUUID(r)}
	claims, err := ClaimsFromRequest(r)
	if err == nil {
		sig.UserID = &claims.UserID
	}
	return sig
}
