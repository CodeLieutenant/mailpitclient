package mailpit_go_api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_GetMessageSource(t *testing.T) {
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
			serverResponse: "Return-Path: <sender@example.com>\nReceived: from...\n\nMessage body",
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:           "message not found",
			messageID:      "nonexistent-id",
			serverResponse: `{"error": "message not found"}`,
			serverStatus:   http.StatusNotFound,
			expectError:    true,
			errorType:      ErrorTypeAPI,
		},
		{
			name:           "server error",
			messageID:      "test-id",
			serverResponse: `{"error": "internal server error"}`,
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
			errorType:      ErrorTypeAPI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Contains(t, r.URL.Path, "/messages/"+tt.messageID+"/source")

				if tt.serverStatus == http.StatusOK {
					w.Header().Set("Content-Type", "text/plain")
				} else {
					w.Header().Set("Content-Type", "application/json")
				}
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

			result, err := c.GetMessageSource(t.Context(), tt.messageID)

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
		})
	}
}

func TestClient_DeleteMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		messageID    string
		errorType    ErrorType
		serverStatus int
		expectError  bool
	}{
		{
			name:         "successful deletion",
			messageID:    "test-message-id",
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "message not found",
			messageID:    "nonexistent-id",
			serverStatus: http.StatusNotFound,
			expectError:  true,
			errorType:    ErrorTypeAPI,
		},
		{
			name:         "server error",
			messageID:    "test-id",
			serverStatus: http.StatusInternalServerError,
			expectError:  true,
			errorType:    ErrorTypeAPI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Contains(t, r.URL.Path, "/messages/"+tt.messageID)

				w.WriteHeader(tt.serverStatus)
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

			err = c.DeleteMessage(t.Context(), tt.messageID)

			if tt.expectError {
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_SearchMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		options        *SearchOptions
		name           string
		query          string
		serverResponse string
		errorType      ErrorType
		serverStatus   int
		expectError    bool
	}{
		{
			name:  "successful search",
			query: "test query",
			options: &SearchOptions{
				Tag:   "important",
				Start: 0,
				Limit: 10,
			},
			serverResponse: `{"total": 2, "messages": [{"ID": "1", "Subject": "Test"}]}`,
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:           "empty query",
			query:          "",
			options:        nil,
			serverResponse: `{"total": 0, "messages": []}`,
			serverStatus:   http.StatusOK,
			expectError:    true,
			errorType:      ErrorTypeValidation,
		},
		{
			name:           "server error",
			query:          "test",
			options:        nil,
			serverResponse: `{"error": "search failed"}`,
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
			errorType:      ErrorTypeAPI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectError && tt.query == "" {
					// Don't expect server to be called for validation errors
					t.Errorf("Server should not be called for validation errors")

					return
				}

				require.Equal(t, http.MethodGet, r.Method)
				require.Contains(t, r.URL.Path, "/search")
				require.Equal(t, tt.query, r.URL.Query().Get("query"))

				if tt.options != nil {
					if tt.options.Tag != "" {
						require.Equal(t, tt.options.Tag, r.URL.Query().Get("tag"))
					}
					if tt.options.Limit > 0 {
						require.Equal(t, "10", r.URL.Query().Get("limit"))
					}
				}

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

			result, err := c.SearchMessages(t.Context(), tt.query, tt.options)

			if tt.expectError {
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.name == "successful search" {
					require.Equal(t, 2, result.Total)
				}
			}
		})
	}
}

func TestClient_MarkMessageRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		messageID    string
		errorType    ErrorType
		serverStatus int
		expectError  bool
	}{
		{
			name:         "successful mark as read",
			messageID:    "test-message-id",
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "message not found",
			messageID:    "nonexistent-id",
			serverStatus: http.StatusNotFound,
			expectError:  true,
			errorType:    ErrorTypeAPI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPut, r.Method)
				require.Contains(t, r.URL.Path, "/messages/"+tt.messageID+"/read")

				w.WriteHeader(tt.serverStatus)
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

			err = c.MarkMessageRead(t.Context(), tt.messageID)

			if tt.expectError {
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_MarkMessageUnread(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		messageID    string
		errorType    ErrorType
		serverStatus int
		expectError  bool
	}{
		{
			name:         "successful mark as unread",
			messageID:    "test-message-id",
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "message not found",
			messageID:    "nonexistent-id",
			serverStatus: http.StatusNotFound,
			expectError:  true,
			errorType:    ErrorTypeAPI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPut, r.Method)
				require.Contains(t, r.URL.Path, "/messages/"+tt.messageID+"/unread")

				w.WriteHeader(tt.serverStatus)
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

			err = c.MarkMessageUnread(t.Context(), tt.messageID)

			if tt.expectError {
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_GetMessageLinkCheck(t *testing.T) {
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
			name:      "successful link check",
			messageID: "test-message-id",
			serverResponse: `{
				"links": [
					{"url": "https://example.com", "status": 200},
					{"url": "https://broken.com", "status": 404, "error": "Not found"}
				]
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:           "message not found",
			messageID:      "nonexistent-id",
			serverResponse: `{"error": "message not found"}`,
			serverStatus:   http.StatusNotFound,
			expectError:    true,
			errorType:      ErrorTypeAPI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Contains(t, r.URL.Path, "/message/"+tt.messageID+"/link-check")

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

			result, err := c.GetMessageLinkCheck(t.Context(), tt.messageID)

			if tt.expectError {
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Len(t, result.Links, 2)
				require.Equal(t, "https://example.com", result.Links[0].URL)
				require.Equal(t, float64(200), result.Links[0].Status) // JSON numbers become float64 with any type
			}
		})
	}
}

func TestClient_GetMessageSpamAssassinCheck(t *testing.T) {
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
			name:      "successful spam check",
			messageID: "test-message-id",
			serverResponse: `{
				"score": 2.5,
				"symbols": [{"name": "TEST_SYMBOL", "score": 1.0, "description": "Test symbol"}],
				"report": [{"score": 1.0, "description": "Test report"}]
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:           "message not found",
			messageID:      "nonexistent-id",
			serverResponse: `{"error": "message not found"}`,
			serverStatus:   http.StatusNotFound,
			expectError:    true,
			errorType:      ErrorTypeAPI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Contains(t, r.URL.Path, "/message/"+tt.messageID+"/sa-check")

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

			result, err := c.GetMessageSpamAssassinCheck(t.Context(), tt.messageID)

			if tt.expectError {
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.InDelta(t, 2.5, result.Score, 0.01)
				require.Len(t, result.Symbols, 1)
				require.Equal(t, "TEST_SYMBOL", result.Symbols[0].Name)
			}
		})
	}
}

func TestClient_DeleteSearchResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		query        string
		errorType    ErrorType
		serverStatus int
		expectError  bool
	}{
		{
			name:         "successful deletion",
			query:        "test query",
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "server error",
			query:        "test",
			serverStatus: http.StatusInternalServerError,
			expectError:  true,
			errorType:    ErrorTypeAPI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Contains(t, r.URL.Path, "/search")
				require.Equal(t, tt.query, r.URL.Query().Get("query"))

				w.WriteHeader(tt.serverStatus)
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

			err = c.DeleteSearchResults(t.Context(), tt.query)

			if tt.expectError {
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
