package mailpit_go_api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_SetTags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		serverResponse string
		errorType      ErrorType
		tags           []string
		serverStatus   int
		expectError    bool
	}{
		{
			name:           "successful request",
			tags:           []string{"important", "work", "personal"},
			serverResponse: `["important", "work", "personal"]`,
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:           "empty tags",
			tags:           []string{},
			serverResponse: `[]`,
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:           "server error",
			tags:           []string{"test"},
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
				// Check method and path
				require.Equal(t, http.MethodPut, r.Method)
				require.True(t, strings.HasSuffix(r.URL.Path, "/tags"))

				// Check headers
				require.Equal(t, "mailpit-go-client/1.0.0", r.Header.Get("User-Agent"))
				require.Equal(t, "application/json", r.Header.Get("Content-Type"))
				require.Equal(t, "application/json", r.Header.Get("Accept"))

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

			result, err := c.SetTags(t.Context(), tt.tags)

			if tt.expectError {
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tt.tags, result)
			}
		})
	}
}

func TestClient_SetMessageTags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		tag          string
		errorType    ErrorType
		messageIDs   []string
		serverStatus int
		expectError  bool
	}{
		{
			name:         "successful request",
			tag:          "important",
			messageIDs:   []string{"msg1", "msg2", "msg3"},
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "empty tag",
			tag:          "",
			messageIDs:   []string{"msg1"},
			serverStatus: http.StatusOK,
			expectError:  true,
			errorType:    ErrorTypeValidation,
		},
		{
			name:         "empty message IDs",
			tag:          "work",
			messageIDs:   []string{},
			serverStatus: http.StatusOK,
			expectError:  true,
			errorType:    ErrorTypeValidation,
		},
		{
			name:         "server error",
			tag:          "test",
			messageIDs:   []string{"msg1"},
			serverStatus: http.StatusInternalServerError,
			expectError:  true,
			errorType:    ErrorTypeAPI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectError && (tt.tag == "" || len(tt.messageIDs) == 0) {
					// Don't expect server to be called for validation errors
					t.Errorf("Server should not be called for validation errors")

					return
				}

				// Check method and path
				require.Equal(t, http.MethodPut, r.Method)
				require.Contains(t, r.URL.Path, "/tags/"+tt.tag)

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

			err = c.SetMessageTags(t.Context(), tt.tag, tt.messageIDs)

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

func TestClient_DeleteTag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		tag          string
		errorType    ErrorType
		serverStatus int
		expectError  bool
	}{
		{
			name:         "successful deletion",
			tag:          "important",
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "tag not found",
			tag:          "nonexistent",
			serverStatus: http.StatusNotFound,
			expectError:  true,
			errorType:    ErrorTypeAPI,
		},
		{
			name:         "empty tag",
			tag:          "",
			serverStatus: http.StatusOK,
			expectError:  true,
			errorType:    ErrorTypeValidation,
		},
		{
			name:         "server error",
			tag:          "work",
			serverStatus: http.StatusInternalServerError,
			expectError:  true,
			errorType:    ErrorTypeAPI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectError && tt.tag == "" {
					// Don't expect server to be called for validation errors
					t.Errorf("Server should not be called for validation errors")

					return
				}

				// Check method and path
				require.Equal(t, http.MethodDelete, r.Method)
				require.Contains(t, r.URL.Path, "/tags/"+tt.tag)

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

			err = c.DeleteTag(t.Context(), tt.tag)

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
