package validation

import (
	"github.com/yerobalg/wealthpulse-service/helper/validator"
)

// Password returns validation messages for a password field.
func Password(field, label string) []validator.Response {
	return []validator.Response{
		validator.Required(field, label),
		validator.Min(field, label, 8),
		validator.PrintASCII(field, label),
		validator.ContainsAny(field, label),
	}
}
