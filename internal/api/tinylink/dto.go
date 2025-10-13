package tinylink

import (
	"net/url"
	"regexp"

	"github.com/Kostaaa1/tinylink/pkg/validator"
)

var (
	aliasRx = regexp.MustCompile("[a-zA-Z0-9]")
)

type CreateTinylinkRequest struct {
	URL     string  `json:"url"`
	Alias   *string `json:"alias"`
	Domain  *string `json:"domain,omitempty"`
	Private bool    `json:"private"`
}

func (r CreateTinylinkRequest) Validate(v *validator.Validator) error {
	_, err := url.Parse(r.URL)
	v.Check(err == nil, "url", "malformed url (parsing failed)")
	v.Check(aliasRx.MatchString(*r.Alias), "alias", "wrong format - use letters and numbers for aliases")
	return nil
}

type UpdateTinylinkRequest struct {
	ID  uint64  `json:"id"`
	URL *string `json:"url"`
	// only authenticated users can update/delete
	UserID  uint64  `json:"user_id"`
	Alias   *string `json:"alias"`
	Domain  *string `json:"domain,omitempty"`
	Private bool    `json:"private"`
}

func (r UpdateTinylinkRequest) Validate(v *validator.Validator) error {
	_, err := url.Parse(*r.URL)
	v.Check(err == nil, "url", "malformed url (parsing failed)")
	v.Check(aliasRx.MatchString(*r.Alias), "alias", "wrong format - use letters and numbers for aliases")
	return nil
}
