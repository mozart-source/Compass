package notification

import "errors"

// Common notification errors
var (
	// ErrNotFound is returned when a notification is not found
	ErrNotFound = errors.New("notification not found")
	// ErrForbidden is returned when a user doesn't have permission to access a notification
	ErrForbidden = errors.New("forbidden")
)

// ErrNotFoundType is a type for notification not found errors
type ErrNotFoundType struct {
	Message string
}

// Error implements the error interface
func (e *ErrNotFoundType) Error() string {
	if e.Message == "" {
		return "notification not found"
	}
	return e.Message
}
