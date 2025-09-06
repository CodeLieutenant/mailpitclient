package mailpit_go_api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_GetServerInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		serverResponse string
		errorType      ErrorType
		serverStatus   int
		expectError    bool
	}{
		{
			name: "successful request",
			serverResponse: `{
				"version": "1.10.0",
				"runtime": "go1.21.0",
				"database": "/tmp/mailpit.db",
				"smtp": 1025,
				"http": 8025,
				"tags": ["test", "work"]
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:           "server error",
			serverResponse: `{"error": "internal server error"}`,
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
			errorType:      ErrorTypeAPI,
		},
		{
			name:           "invalid JSON response",
			serverResponse: `invalid json`,
			serverStatus:   http.StatusOK,
			expectError:    true,
			errorType:      ErrorTypeResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check method and path
				require.Equal(t, http.MethodGet, r.Method)
				require.True(t, strings.HasSuffix(r.URL.Path, "/info"))

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

			result, err := c.GetServerInfo(t.Context())

			if tt.expectError {
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, "1.10.0", result.Version)
				require.Equal(t, "go1.21.0", result.Runtime)
				require.Equal(t, "/tmp/mailpit.db", result.Database)
				require.Equal(t, 1025, result.SMTPPort)
				require.Equal(t, 8025, result.HTTPPort)
				require.Equal(t, []string{"test", "work"}, result.Tags)
			}
		})
	}
}

func TestClient_GetStats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		serverResponse string
		errorType      ErrorType
		serverStatus   int
		expectError    bool
	}{
		{
			name: "successful request",
			serverResponse: `{
				"total": 100,
				"unread": 25,
				"tags": ["important", "work", "personal"]
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:           "server error",
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
				require.True(t, strings.HasSuffix(r.URL.Path, "/stats"))

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

			result, err := c.GetStats(t.Context())

			if tt.expectError {
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, 100, result.Total)
				require.Equal(t, 25, result.Unread)
				require.Equal(t, []string{"important", "work", "personal"}, result.Tags)
			}
		})
	}
}

func TestClient_GetTags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		serverResponse string
		errorType      ErrorType
		serverStatus   int
		expectError    bool
	}{
		{
			name:           "successful request with tags",
			serverResponse: `["important", "work", "personal", "test"]`,
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:           "successful request empty tags",
			serverResponse: `[]`,
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:           "server error",
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
				require.True(t, strings.HasSuffix(r.URL.Path, "/tags"))

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

			result, err := c.GetTags(t.Context())

			if tt.expectError {
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				switch tt.name {
				case "successful request with tags":
					expected := []string{"important", "work", "personal", "test"}
					require.Equal(t, expected, result)
				case "successful request empty tags":
					require.Empty(t, result)
				default:
				}
			}
		})
	}
}

func TestClient_Ping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		errorType    ErrorType
		serverStatus int
		expectError  bool
	}{
		{
			name:         "successful ping",
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "server error",
			serverStatus: http.StatusInternalServerError,
			expectError:  true,
			errorType:    ErrorTypeAPI,
		},
		{
			name:         "not found",
			serverStatus: http.StatusNotFound,
			expectError:  true,
			errorType:    ErrorTypeAPI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodHead, r.Method)
				require.True(t, strings.HasSuffix(r.URL.Path, "/info"))

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

			err = c.Ping(t.Context())

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

func TestClient_GetWebUIConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		serverResponse string
		errorType      ErrorType
		serverStatus   int
		expectError    bool
	}{
		{
			name: "successful request",
			serverResponse: `{
				"ReadOnly": false,
				"Version": "v1.10.0",
				"ShowVersions": true
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:           "server error",
			serverResponse: `{"error": "internal server error"}`,
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
			errorType:      ErrorTypeAPI,
		},
		{
			name:           "invalid JSON response",
			serverResponse: `invalid json`,
			serverStatus:   http.StatusOK,
			expectError:    true,
			errorType:      ErrorTypeResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.True(t, strings.HasSuffix(r.URL.Path, "/webui"))

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

			result, err := c.GetWebUIConfig(t.Context())

			if tt.expectError {
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.False(t, result.ReadOnly)
				require.Equal(t, "v1.10.0", result.Version)
				require.True(t, result.ShowVersions)
			}
		})
	}
}
