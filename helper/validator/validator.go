package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"

	errorLib "github.com/yerobalg/wealthpulse-service/helper/errors"
)

var regexCache sync.Map

func compileRegex(pattern string) *regexp.Regexp {
	if cached, ok := regexCache.Load(pattern); ok {
		return cached.(*regexp.Regexp)
	}
	re := regexp.MustCompile(pattern)
	regexCache.Store(pattern, re)
	return re
}

type validatorLib struct {
	validator *validator.Validate
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Response struct {
	Tag     string
	Field   string
	Message string
}

type Interface interface {
	ValidateStruct(any) error
	GetFieldErrors(err error, responses []Response) (string, []FieldError)
	Bind(req any, messages []Response) error
}

func Init() Interface {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("json")
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})

	v.RegisterValidation("regex", func(fl validator.FieldLevel) bool {
		return compileRegex(fl.Param()).MatchString(fl.Field().String())
	})

	return &validatorLib{validator: v}
}

func (v *validatorLib) ValidateStruct(data any) error {
	return v.validator.Struct(data)
}

func (v *validatorLib) GetFieldErrors(err error, responses []Response) (string, []FieldError) {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return "", nil
	}

	messageMap := make(map[string]string)
	for _, response := range responses {
		key := response.Field + "." + response.Tag
		messageMap[key] = response.Message
	}

	fieldErrors := make([]FieldError, len(validationErrors))
	for i, fe := range validationErrors {
		message := fmt.Sprintf("Field '%s' failed on '%s' validation", fe.Field(), fe.Tag())

		key := fe.Field() + "." + fe.Tag()
		if custom, exists := messageMap[key]; exists {
			message = custom
		}

		fieldErrors[i] = FieldError{
			Field:   fe.Field(),
			Message: message,
		}
	}

	errSummary := fmt.Sprintf("Terdapat %d kesalahan pada data yang dikirim", len(fieldErrors))

	return errSummary, fieldErrors
}

func (v *validatorLib) Bind(req any, messages []Response) error {
	if err := v.ValidateStruct(req); err != nil {
		summary, fields := v.GetFieldErrors(err, messages)
		return errorLib.BadRequest(summary, fields)
	}
	return nil
}

// --- Single-tag message helpers ---
// Each returns one Response for a specific validation tag.

func Required(field, label string) Response {
	return Response{Tag: "required", Field: field, Message: label + " wajib diisi"}
}

func GT(field, label string) Response {
	return Response{Tag: "gt", Field: field, Message: label + " tidak valid"}
}

func GTE(field, label string) Response {
	return Response{Tag: "gte", Field: field, Message: label + " tidak valid"}
}

func Max(field, label string, n int) Response {
	return Response{Tag: "max", Field: field, Message: fmt.Sprintf("%s maksimal %d karakter", label, n)}
}

func MaxDigit(field, label string, n int) Response {
	return Response{Tag: "max", Field: field, Message: fmt.Sprintf("%s maksimal %d digit", label, n)}
}

func Min(field, label string, n int) Response {
	return Response{Tag: "min", Field: field, Message: fmt.Sprintf("%s minimal %d karakter", label, n)}
}

func Len(field, label string, n int) Response {
	return Response{Tag: "len", Field: field, Message: fmt.Sprintf("%s harus %d digit", label, n)}
}

func LTE(field, label string, n int) Response {
	return Response{Tag: "lte", Field: field, Message: fmt.Sprintf("%s maksimal %d", label, n)}
}

func Numeric(field, label string) Response {
	return Response{Tag: "numeric", Field: field, Message: label + " hanya boleh berupa angka"}
}

func PrintASCII(field, label string) Response {
	return Response{Tag: "printascii", Field: field, Message: label + " hanya boleh mengandung karakter yang dapat dicetak"}
}

func ContainsAny(field, label string) Response {
	return Response{Tag: "containsany", Field: field, Message: label + " harus mengandung kombinasi huruf dan angka"}
}

func Email(field, label string) Response {
	return Response{Tag: "email", Field: field, Message: "Format " + strings.ToLower(label) + " tidak valid"}
}

// Regex accepts a fully custom message since regex error wording is always field-specific.
func Regex(field, message string) Response {
	return Response{Tag: "regex", Field: field, Message: message}
}

// IsColor accepts a fully custom message.
func IsColor(field, message string) Response {
	return Response{Tag: "iscolor", Field: field, Message: message}
}

// RequiredIf accepts a fully custom message.
func RequiredIf(field, message string) Response {
	return Response{Tag: "required_if", Field: field, Message: message}
}

// --- Concat ---

// Concat merges multiple Response slices into one.
func Concat(slices ...[]Response) []Response {
	total := 0
	for _, s := range slices {
		total += len(s)
	}
	result := make([]Response, 0, total)
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

// --- Compound builders (convenience for common multi-tag shapes) ---

// RequiredString returns required + max-length messages for a string field.
func RequiredString(field, label string, max int) []Response {
	return []Response{
		Required(field, label),
		Max(field, label, max),
	}
}

// RequiredID returns required + gt=0 messages for an ID field.
func RequiredID(field, label string) []Response {
	return []Response{
		Required(field, label),
		GT(field, label),
	}
}

// RequiredDigits returns required + numeric + max-length messages for a digits-only field.
func RequiredDigits(field, label string, max int) []Response {
	return []Response{
		Required(field, label),
		Numeric(field, label),
		MaxDigit(field, label, max),
	}
}

// RequiredEmail returns required + email format + max-length messages.
func RequiredEmail(field, label string) []Response {
	return []Response{
		Required(field, label),
		Email(field, label),
		Max(field, label, 255),
	}
}

// RequiredUsername returns required + max-length + no-spaces messages for a username field.
func RequiredUsername(field string) []Response {
	return []Response{
		Required(field, "Username"),
		Max(field, "Username", 255),
		Regex(field, "Username tidak boleh mengandung spasi"),
	}
}
