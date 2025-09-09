package mailpitclient

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name: "error with status code",
			err: &Error{
				Type:       ErrorTypeAPI,
				Message:    "resource not found",
				StatusCode: 404,
			},
			expected: "mailpit api error (status 404): resource not found",
		},
		{
			name: "error without status code",
			err: &Error{
				Type:    ErrorTypeNetwork,
				Message: "connection failed",
			},
			expected: "mailpit network error: connection failed",
		},
		{
			name: "config error",
			err: &Error{
				Type:    ErrorTypeConfig,
				Message: "invalid base URL",
			},
			expected: "mailpit config error: invalid base URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	t.Parallel()

	originalErr := errors.New("original error") // nolint:err113

	tests := []struct {
		expectedErr error
		err         *Error
		name        string
	}{
		{
			name: "error with cause",
			err: &Error{
				Type:    ErrorTypeRequest,
				Message: "request failed",
				Cause:   originalErr,
			},
			expectedErr: originalErr,
		},
		{
			name: "error without cause",
			err: &Error{
				Type:    ErrorTypeValidation,
				Message: "validation failed",
				Cause:   nil,
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.expectedErr, tt.err.Unwrap())
		})
	}
}

func TestError_IsType(t *testing.T) {
	t.Parallel()

	err := &Error{
		Type:    ErrorTypeValidation,
		Message: "validation failed",
	}

	tests := []struct {
		name      string
		errorType ErrorType
		expected  bool
	}{
		{
			name:      "matching type",
			errorType: ErrorTypeValidation,
			expected:  true,
		},
		{
			name:      "non-matching type",
			errorType: ErrorTypeNetwork,
			expected:  false,
		},
		{
			name:      "another non-matching type",
			errorType: ErrorTypeAPI,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.expected, err.IsType(tt.errorType))
		})
	}
}

func TestError_IsAPIError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err        *Error
		name       string
		statusCode int
		expected   bool
	}{
		{
			name: "matching API error with status code",
			err: &Error{
				Type:       ErrorTypeAPI,
				Message:    "not found",
				StatusCode: 404,
			},
			statusCode: 404,
			expected:   true,
		},
		{
			name: "API error with different status code",
			err: &Error{
				Type:       ErrorTypeAPI,
				Message:    "not found",
				StatusCode: 404,
			},
			statusCode: 500,
			expected:   false,
		},
		{
			name: "non-API error",
			err: &Error{
				Type:       ErrorTypeNetwork,
				Message:    "connection failed",
				StatusCode: 500,
			},
			statusCode: 500,
			expected:   false,
		},
		{
			name: "API error without status code",
			err: &Error{
				Type:    ErrorTypeAPI,
				Message: "general error",
			},
			statusCode: 500,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.expected, tt.err.IsAPIError(tt.statusCode))
		})
	}
}

func TestNewConfigError(t *testing.T) {
	t.Parallel()

	message := "invalid configuration"
	err := NewConfigError(message)

	require.NotNil(t, err)
	require.Equal(t, ErrorTypeConfig, err.Type)
	require.Equal(t, message, err.Message)
	require.NoError(t, err.Cause)
	require.Equal(t, 0, err.StatusCode)
	require.Empty(t, err.Response)
}

func TestNewValidationError(t *testing.T) {
	t.Parallel()

	message := "validation failed"
	err := NewValidationError(message)

	require.NotNil(t, err)
	require.Equal(t, ErrorTypeValidation, err.Type)
	require.Equal(t, message, err.Message)
	require.NoError(t, err.Cause)
	require.Equal(t, 0, err.StatusCode)
	require.Empty(t, err.Response)
}

func TestErrorTypes(t *testing.T) {
	t.Parallel()

	// Test that all error types are defined as expected
	require.Equal(t, ErrorTypeConfig, ErrorType("config"))
	require.Equal(t, ErrorTypeNetwork, ErrorType("network"))
	require.Equal(t, ErrorTypeRequest, ErrorType("request"))
	require.Equal(t, ErrorTypeResponse, ErrorType("response"))
	require.Equal(t, ErrorTypeAPI, ErrorType("api"))
	require.Equal(t, ErrorTypeValidation, ErrorType("validation"))
}

func TestError_ErrorInterface(t *testing.T) {
	t.Parallel()

	// Test that Error implements the error interface
	var err error = &Error{
		Type:    ErrorTypeNetwork,
		Message: "test error",
	}

	require.Equal(t, "mailpit network error: test error", err.Error())
}

func TestError_WrappingInterface(t *testing.T) {
	t.Parallel()

	// Test that Error works with errors.Unwrap and errors.Is
	originalErr := errors.New("original error") // nolint:err113

	mailpitErr := &Error{
		Type:    ErrorTypeRequest,
		Message: "request failed",
		Cause:   originalErr,
	}

	// Test that errors.Unwrap works
	unwrappedErr := errors.Unwrap(mailpitErr)
	require.Equal(t, originalErr, unwrappedErr)

	// Test that errors.Is works
	require.ErrorIs(t, mailpitErr, originalErr)
}
