package errors

import (
	"net/http"
	"reflect"
)

const (
	TypeOK                  = "OK"
	TypeBadRequest          = "Bad Request"
	TypeNotFound            = "Not Found"
	TypeInternalServerError = "Internal Server Error"
	TypeUnauthorized        = "Unauthorized"
)

// Errors represents a structured application error with an HTTP status code,
// a human-readable message, and a type string derived from the HTTP status text.
type Errors struct {
	Type    string
	Code    int64
	Message string
	Data    any
}

func (e *Errors) Error() string {
	return e.Message
}

// NewWithCode creates a new Errors instance with the given HTTP status code,
// message, and type string.
func NewWithCode(code int64, message, errType string) error {
	errors := &Errors{
		Type:    errType,
		Code:    code,
		Message: message,
	}

	return errors
}

// NotFound returns a 404 Not Found error with a message in the format "<entity> not found".
func NotFound(entity string) error {
	return NewWithCode(http.StatusNotFound, entity+" not found", TypeNotFound)
}

// InternalServerError returns a 500 Internal Server Error with the given message.
func InternalServerError(message string) error {
	if message == "" {
		message = "Something went wrong"
	}
	return NewWithCode(http.StatusInternalServerError, message, TypeInternalServerError)
}

// BadRequest returns a 400 Bad Request error with the given message and data.
func BadRequest(message string, data any) error {
	err := &Errors{
		Type:    TypeBadRequest,
		Code:    http.StatusBadRequest,
		Message: message,
		Data:    data,
	}
	return err
}

func Unauthorized(message string) error {
	return NewWithCode(http.StatusUnauthorized, message, TypeUnauthorized)
}

// IsApplicationError checks whether the given error is a known application error.
func IsApplicationError(err error) bool {
	return reflect.TypeOf(err).String() == "*errors.Errors"
}

// GetType returns the error type string of the given error.
// Returns http.StatusText(200) if err is nil, the Errors.Type if it is a known
// application error, or http.StatusText(500) for any other error type.
func GetType(err error) string {
	if err == nil {
		return TypeOK
	}

	if IsApplicationError(err) {
		return err.(*Errors).Type
	}

	return TypeInternalServerError
}

// GetCode returns the HTTP status code associated with the given error.
// Returns 200 if err is nil, the Errors.Code if it is a known application error,
// or 500 for any other error type.
func GetCode(err error) int64 {
	if err == nil {
		return 200
	}

	if IsApplicationError(err) {
		return err.(*Errors).Code
	}

	return 500
}

// GetMessage returns the message associated with the given error.
// Returns "OK" if err is nil, the Errors.Message if it is a known application error,
// or the raw error string for any other error type.
func GetMessage(err error) string {
	if err == nil {
		return "OK"
	}

	if IsApplicationError(err) {
		return err.(*Errors).Message
	}

	return err.Error()
}

// GetData returns the data associated with the given error.
// Returns nil if err is nil or not a known application error.
func GetData(err error) any {
	if err == nil {
		return nil
	}

	if IsApplicationError(err) {
		return err.(*Errors).Data
	}

	return nil
}

// GetAll extracts all error information (type, code, message, and data) into a map.
func GetAll(err error) map[string]any {
	return map[string]any{
		"type":    GetType(err),
		"code":    GetCode(err),
		"message": GetMessage(err),
		"data":    GetData(err),
	}
}

// Is checks whether the given error matches the specified error type.
// Returns false if err is not nil, true if err is a known application error
// whose Type matches errType, or true if errType is TypeInternalServerError
// for any other unrecognized error.
func Is(err error, errType string) bool {
	if err == nil {
		return false
	}

	if IsApplicationError(err) {
		return err.(*Errors).Type == errType
	}

	return errType == TypeInternalServerError
}
