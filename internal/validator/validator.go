package validator

import "regexp"

type Validator struct {
	Errors map[string]string
}

var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, value string) {
	if _, ok := v.Errors[key]; !ok {
		v.Errors[key] = value
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func In(value string, list []string) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}
	return false
}

func Unique(value []string) bool {
	uniq := make(map[string]struct{})
	for _, v := range value {
		if _, ok := uniq[v]; !ok {
			uniq[v] = struct{}{}
		}
	}
	return len(uniq) == len(value)
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
