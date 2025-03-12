package request

import (
	"net/url"

	"github.com/Kostaaa1/tinylink/internal/validator"
)

type CreateTinylinkRequest struct {
	URL    string `json:"url"`
	Alias  string `json:"alias"`
	Domain string `json:"domain"`
}

func (req *CreateTinylinkRequest) IsValid(v *validator.Validator) bool {
	v.Check(req.URL != "", "url", "must be provided")
	_, err := url.ParseRequestURI(req.URL)
	v.Check(err == nil, "url", "invalid url format")
	v.Check(!(req.Alias != "" && len(req.Alias) < 5), "alias", "must be at least 5 characters long")
	return v.Valid()
}
