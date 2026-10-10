package utils

import "errors"

var ErrUserExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")

// ErrorResponse is the app-level HTTP error (FieldOS ErrorResponse).
type ErrorResponse struct {
	Message string
	Status  int
	Errors  []string
	Data    any
}

func NewErrorResponse(message string, status int, errs []string, data any) *ErrorResponse {
	if errs == nil {
		errs = []string{}
	}
	return &ErrorResponse{
		Message: message,
		Status:  status,
		Errors:  errs,
		Data:    data,
	}
}

func (e *ErrorResponse) Error() string {
	return e.Message
}

// ErrorsAsAny converts string error details for JSON response DTOs.
func ErrorsAsAny(errs []string) []any {
	out := make([]any, 0, len(errs))
	for _, e := range errs {
		out = append(out, e)
	}
	return out
}
