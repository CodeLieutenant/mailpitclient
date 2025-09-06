package mailpit_go_api

import "fmt"

// ErrorType represents the type of error that occurred.
type ErrorType string

const (
	// ErrorTypeConfig indicates a configuration error
	ErrorTypeConfig ErrorType = "config"

	// ErrorTypeNetwork indicates a network-related error
	ErrorTypeNetwork ErrorType = "network"

	// ErrorTypeRequest indicates an error creating the HTTP request
	ErrorTypeRequest ErrorType = "request"

	// ErrorTypeResponse indicates an error parsing the HTTP response
	ErrorTypeResponse ErrorType = "response"

	// ErrorTypeAPI indicates an API error returned by the server
	ErrorTypeAPI ErrorType = "api"

	// ErrorTypeValidation indicates a validation error
	ErrorTypeValidation ErrorType = "validation"
)

// Error represents a Mailpit client error with structured information.
type Error struct {
	Cause      error     `json:"-"`
	Type       ErrorType `json:"type"`
	Message    string    `json:"message"`
	Response   string    `json:"response,omitempty"`
	StatusCode int       `json:"status_code,omitempty"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("mailpit %s error (status %d): %s", e.Type, e.StatusCode, e.Message)
	}

	return fmt.Sprintf("mailpit %s error: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	return e.Cause
}

// IsType checks if the error is of a specific type.
func (e *Error) IsType(errorType ErrorType) bool {
	return e.Type == errorType
}

// IsAPIError checks if the error is an API error with the given status code.
func (e *Error) IsAPIError(statusCode int) bool {
	return e.Type == ErrorTypeAPI && e.StatusCode == statusCode
}

// NewConfigError creates a new configuration error.
func NewConfigError(message string) *Error {
	return &Error{
		Type:    ErrorTypeConfig,
		Message: message,
	}
}

// NewValidationError creates a new validation error.
func NewValidationError(message string) *Error {
	return &Error{
		Type:    ErrorTypeValidation,
		Message: message,
	}
}
