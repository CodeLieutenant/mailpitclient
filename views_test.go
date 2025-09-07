package mailpit_go_api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_GetMessageRaw(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		messageID      string
		serverResponse string
		errorType      ErrorType
		serverStatus   int
		expectError    bool
	}{
		{
			name:           "successful request",
			messageID:      "test-message-id",
			serverStatus:   http.StatusOK,
			serverResponse: "From: sender@example.com\r\nTo: recipient@example.com\r\nSubject: Test\r\n\r\nTest message body",
			expectError:    false,
		},
		{
			name:        "empty message ID",
			messageID:   "",
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:         "message not found",
			messageID:    "nonexistent",
			serverStatus: http.StatusNotFound,
			expectError:  true,
			errorType:    ErrorTypeAPI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.messageID != "" {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodGet, r.Method)
					require.True(t, strings.HasSuffix(r.URL.Path, "/view/"+tt.messageID+".raw"))

					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(tt.serverStatus)
					_, _ = w.Write([]byte(tt.serverResponse))
				}))
				defer server.Close()

				config := &Config{
					BaseURL:    server.URL,
					APIPath:    "/api/v1",
					MaxRetries: 0,
					HTTPClient: &http.Client{Timeout: 5 * time.Second},
				}

				c, err := NewClient(config)
				require.NoError(t, err)
				defer c.Close()

				result, err := c.GetMessageRaw(t.Context(), tt.messageID)

				if tt.expectError {
					require.Error(t, err)
					var mailpitErr *Error
					require.ErrorAs(t, err, &mailpitErr)
					require.Equal(t, tt.errorType, mailpitErr.Type)
					require.Empty(t, result)
				} else {
					require.NoError(t, err)
					require.Equal(t, tt.serverResponse, result)
				}
			} else {
				// Test empty message ID without server
				config := &Config{
					BaseURL:    "http://localhost:8025",
					APIPath:    "/api/v1",
					HTTPClient: &http.Client{Timeout: 5 * time.Second},
				}

				c, err := NewClient(config)
				require.NoError(t, err)
				defer c.Close()

				result, err := c.GetMessageRaw(t.Context(), tt.messageID)
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
				require.Empty(t, result)
			}
		})
	}
}

func TestClient_GetMessagePartHTML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		messageID      string
		partID         string
		serverResponse string
		errorType      ErrorType
		serverStatus   int
		expectError    bool
	}{
		{
			name:           "successful request",
			messageID:      "test-message-id",
			partID:         "part-1",
			serverStatus:   http.StatusOK,
			serverResponse: "<html><body>Test HTML content</body></html>",
			expectError:    false,
		},
		{
			name:        "empty message ID",
			messageID:   "",
			partID:      "part-1",
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "empty part ID",
			messageID:   "test-message-id",
			partID:      "",
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:         "part not found",
			messageID:    "test-message-id",
			partID:       "nonexistent",
			serverStatus: http.StatusNotFound,
			expectError:  true,
			errorType:    ErrorTypeAPI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.messageID != "" && tt.partID != "" {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodGet, r.Method)
					expectedPath := "/view/" + tt.messageID + "/part/" + tt.partID + ".html"
					require.True(t, strings.HasSuffix(r.URL.Path, expectedPath))

					w.Header().Set("Content-Type", "text/html")
					w.WriteHeader(tt.serverStatus)
					_, _ = w.Write([]byte(tt.serverResponse))
				}))
				defer server.Close()

				config := &Config{
					BaseURL:    server.URL,
					APIPath:    "/api/v1",
					MaxRetries: 0,
					HTTPClient: &http.Client{Timeout: 5 * time.Second},
				}

				c, err := NewClient(config)
				require.NoError(t, err)
				defer c.Close()

				result, err := c.GetMessagePartHTML(t.Context(), tt.messageID, tt.partID)

				if tt.expectError {
					require.Error(t, err)
					var mailpitErr *Error
					require.ErrorAs(t, err, &mailpitErr)
					require.Equal(t, tt.errorType, mailpitErr.Type)
					require.Empty(t, result)
				} else {
					require.NoError(t, err)
					require.Equal(t, tt.serverResponse, result)
				}
			} else {
				// Test validation errors without server
				config := &Config{
					BaseURL:    "http://localhost:8025",
					APIPath:    "/api/v1",
					HTTPClient: &http.Client{Timeout: 5 * time.Second},
				}

				c, err := NewClient(config)
				require.NoError(t, err)
				defer c.Close()

				result, err := c.GetMessagePartHTML(t.Context(), tt.messageID, tt.partID)
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
				require.Empty(t, result)
			}
		})
	}
}

func TestClient_GetMessagePartText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		messageID      string
		partID         string
		serverResponse string
		errorType      ErrorType
		serverStatus   int
		expectError    bool
	}{
		{
			name:           "successful request",
			messageID:      "test-message-id",
			partID:         "part-1",
			serverStatus:   http.StatusOK,
			serverResponse: "Test text content",
			expectError:    false,
		},
		{
			name:        "empty message ID",
			messageID:   "",
			partID:      "part-1",
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "empty part ID",
			messageID:   "test-message-id",
			partID:      "",
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.messageID != "" && tt.partID != "" {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodGet, r.Method)
					expectedPath := "/view/" + tt.messageID + "/part/" + tt.partID + ".text"
					require.True(t, strings.HasSuffix(r.URL.Path, expectedPath))

					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(tt.serverStatus)
					_, _ = w.Write([]byte(tt.serverResponse))
				}))
				defer server.Close()

				config := &Config{
					BaseURL:    server.URL,
					APIPath:    "/api/v1",
					MaxRetries: 0,
					HTTPClient: &http.Client{Timeout: 5 * time.Second},
				}

				c, err := NewClient(config)
				require.NoError(t, err)
				defer c.Close()

				result, err := c.GetMessagePartText(t.Context(), tt.messageID, tt.partID)

				if tt.expectError {
					require.Error(t, err)
					var mailpitErr *Error
					require.ErrorAs(t, err, &mailpitErr)
					require.Equal(t, tt.errorType, mailpitErr.Type)
					require.Empty(t, result)
				} else {
					require.NoError(t, err)
					require.Equal(t, tt.serverResponse, result)
				}
			} else {
				// Test validation errors without server
				config := &Config{
					BaseURL:    "http://localhost:8025",
					APIPath:    "/api/v1",
					HTTPClient: &http.Client{Timeout: 5 * time.Second},
				}

				c, err := NewClient(config)
				require.NoError(t, err)
				defer c.Close()

				result, err := c.GetMessagePartText(t.Context(), tt.messageID, tt.partID)
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
				require.Empty(t, result)
			}
		})
	}
}

func TestClient_GetMessageEvents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		messageID      string
		serverResponse string
		errorType      ErrorType
		serverStatus   int
		expectError    bool
	}{
		{
			name:         "successful request",
			messageID:    "test-message-id",
			serverStatus: http.StatusOK,
			serverResponse: `{
				"events": [
					{
						"ID": "event-1",
						"Type": "delivery",
						"Timestamp": "2023-01-01T12:00:00Z",
						"Data": {"status": "sent"}
					}
				]
			}`,
			expectError: false,
		},
		{
			name:        "empty message ID",
			messageID:   "",
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:         "message not found",
			messageID:    "nonexistent",
			serverStatus: http.StatusNotFound,
			expectError:  true,
			errorType:    ErrorTypeAPI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.messageID != "" {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodGet, r.Method)
					require.True(t, strings.HasSuffix(r.URL.Path, "/message/"+tt.messageID+"/events"))

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.serverStatus)
					_, _ = w.Write([]byte(tt.serverResponse))
				}))
				defer server.Close()

				config := &Config{
					BaseURL:    server.URL,
					APIPath:    "/api/v1",
					MaxRetries: 0,
					HTTPClient: &http.Client{Timeout: 5 * time.Second},
				}

				c, err := NewClient(config)
				require.NoError(t, err)
				defer c.Close()

				result, err := c.GetMessageEvents(t.Context(), tt.messageID)

				if tt.expectError {
					require.Error(t, err)
					var mailpitErr *Error
					require.ErrorAs(t, err, &mailpitErr)
					require.Equal(t, tt.errorType, mailpitErr.Type)
					require.Nil(t, result)
				} else {
					require.NoError(t, err)
					require.NotNil(t, result)
					require.Len(t, result.Events, 1)
					require.Equal(t, "event-1", result.Events[0].ID)
					require.Equal(t, "delivery", result.Events[0].Type)
				}
			} else {
				// Test empty message ID without server
				config := &Config{
					BaseURL:    "http://localhost:8025",
					APIPath:    "/api/v1",
					HTTPClient: &http.Client{Timeout: 5 * time.Second},
				}

				c, err := NewClient(config)
				require.NoError(t, err)
				defer c.Close()

				result, err := c.GetMessageEvents(t.Context(), tt.messageID)
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
				require.Nil(t, result)
			}
		})
	}
}
