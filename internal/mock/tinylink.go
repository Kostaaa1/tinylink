package mock

import (
	"fmt"
	"math/rand/v2"

	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	"github.com/google/uuid"
)

func GuestTinylink(uuid string) *tinylink.Tinylink {
	val := rand.IntN(10000)
	return &tinylink.Tinylink{
		GuestUUID: uuid,
		Domain:    fmt.Sprintf("domain%d.com", val),
		Alias:     fmt.Sprintf("alias_%d", val),
		URL:       fmt.Sprintf("http://example.com/%d", val),
	}
}

func UserTinylink(userID uint64) *tinylink.Tinylink {
	val := rand.IntN(10000)
	tl := &tinylink.Tinylink{
		UserID:    &userID,
		Alias:     fmt.Sprintf("alias_%d", val),
		URL:       fmt.Sprintf("http://example.com/%d", val),
		Domain:    fmt.Sprintf("domain%d.com", val),
		Private:   rand.IntN(2) == 1,
		GuestUUID: uuid.NewString(),
	}
	return tl
}

func UpdateTinylinkParams(id uint64, userID uint64) tinylink.UpdateTinylinkParams {
	val := rand.IntN(10000)
	alias := fmt.Sprintf("alias_%d", val)
	domain := fmt.Sprintf("domain%d.com", val)
	url := fmt.Sprintf("http://example.com/%d", val)

	return tinylink.UpdateTinylinkParams{
		ID:      id,
		UserID:  userID,
		Alias:   &alias,
		Domain:  &domain,
		URL:     &url,
		Private: rand.IntN(2) == 1,
	}
}

func CreateTinylinkParams(userID *uint64, guestUUID *string) tinylink.CreateTinylinkParams {
	val := rand.IntN(10000)
	alias := fmt.Sprintf("alias_%d", val)
	domain := fmt.Sprintf("domain%d.com", val)

	params := tinylink.CreateTinylinkParams{
		GuestUUID: uuid.NewString(),
		Alias:     &alias,
		Domain:    &domain,
		URL:       fmt.Sprintf("http://example.com/%d", val),
		Private:   rand.IntN(2) == 1,
	}

	if userID != nil {
		params.UserID = userID
	}

	if guestUUID != nil {
		params.GuestUUID = *guestUUID
	}

	return params
}
