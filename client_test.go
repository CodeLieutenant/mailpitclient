package mailpit_go_api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		config      *Config
		name        string
		errorType   ErrorType
		expectError bool
	}{
		{
			name:        "nil config uses defaults",
			config:      nil,
			expectError: false,
		},
		{
			name: "valid config",
			config: &Config{
				BaseURL: "http://localhost:8025",
				Timeout: 30 * time.Second,
			},
			expectError: false,
		},
		{
			name: "empty BaseURL",
			config: &Config{
				BaseURL: "",
			},
			expectError: true,
			errorType:   ErrorTypeConfig,
		},
		{
			name: "invalid BaseURL",
			config: &Config{
				BaseURL: "://invalid-url",
			},
			expectError: true,
			errorType:   ErrorTypeConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c, err := NewClient(tt.config)

			if tt.expectError {
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
				require.Nil(t, c)
			} else {
				require.NoError(t, err)
				require.NotNil(t, c)
				require.NoError(t, c.Close())
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()

	require.Equal(t, "http://localhost:8025", config.BaseURL)
	require.Equal(t, "/api/v1", config.APIPath)
	require.Equal(t, 30*time.Second, config.Timeout)
	require.Equal(t, "mailpit-go-client/1.0.0", config.UserAgent)
	require.Equal(t, 3, config.MaxRetries)
	require.Equal(t, 1*time.Second, config.RetryDelay)
	require.NotNil(t, config.HTTPClient)
}

func TestClient_ListMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		options        *ListOptions
		name           string
		serverResponse string
		errorType      ErrorType
		serverStatus   int
		expectError    bool
	}{
		{
			name:           "successful request",
			serverResponse: `{"total": 2, "messages": [{"ID": "1", "Subject": "Test"}]}`,
			serverStatus:   http.StatusOK,
			options:        nil,
			expectError:    false,
		},
		{
			name:           "with pagination options",
			serverResponse: `{"total": 10, "messages": []}`,
			serverStatus:   http.StatusOK,
			options:        &ListOptions{Start: 0, Limit: 5},
			expectError:    false,
		},
		{
			name:           "server error",
			serverResponse: `{"error": "internal server error"}`,
			serverStatus:   http.StatusInternalServerError,
			options:        nil,
			expectError:    true,
			errorType:      ErrorTypeAPI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check headers
				require.Equal(t, "mailpit-go-client/1.0.0", r.Header.Get("User-Agent"))
				require.Equal(t, "application/json", r.Header.Get("Content-Type"))
				require.Equal(t, "application/json", r.Header.Get("Accept"))

				// Check URL path
				require.True(t, strings.HasSuffix(r.URL.Path, "/messages"))

				// Check query parameters if options provided
				if tt.options != nil {
					query := r.URL.Query()
					if tt.options.Start > 0 {
						require.Equal(t, "0", query.Get("start"))
					}
					if tt.options.Limit > 0 {
						require.Equal(t, "5", query.Get("limit"))
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
				MaxRetries: 0, // No retries for simpler testing
				HTTPClient: &http.Client{Timeout: 5 * time.Second},
			}

			c, err := NewClient(config)
			require.NoError(t, err)
			defer c.Close()

			result, err := c.ListMessages(t.Context(), tt.options)

			if tt.expectError {
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.name == "successful request" {
					require.Equal(t, 2, result.Total)
					require.Len(t, result.Messages, 1)
					require.Equal(t, "1", result.Messages[0].ID)
				}
			}
		})
	}
}

func TestClient_Authentication(t *testing.T) {
	t.Parallel()

	tests := []struct {
		config    *Config
		checkAuth func(r *http.Request) bool
		name      string
	}{
		{
			name: "API key authentication",
			config: &Config{
				BaseURL: "http://localhost:8025",
				APIKey:  "test-api-key",
			},
			checkAuth: func(r *http.Request) bool {
				return r.Header.Get("Authorization") == "Bearer test-api-key"
			},
		},
		{
			name: "basic authentication",
			config: &Config{
				BaseURL:  "http://localhost:8025",
				Username: "testuser",
				Password: "testpass",
			},
			checkAuth: func(r *http.Request) bool {
				username, password, ok := r.BasicAuth()

				return ok && username == "testuser" && password == "testpass"
			},
		},
		{
			name: "no authentication",
			config: &Config{
				BaseURL: "http://localhost:8025",
			},
			checkAuth: func(r *http.Request) bool {
				return r.Header.Get("Authorization") == ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.True(t, tt.checkAuth(r), "Authentication check failed")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"version": "test", "runtime": "go"}`))
			}))
			defer server.Close()

			tt.config.BaseURL = server.URL
			tt.config.APIPath = "/api/v1"

			c, err := NewClient(tt.config)
			require.NoError(t, err)
			defer c.Close()

			// Test authentication by calling a method that will trigger HTTP request
			_, err = c.GetServerInfo(t.Context())
			require.NoError(t, err)
		})
	}
}

func TestClient_GetMessage(t *testing.T) {
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
			serverResponse: `{"ID": "test-message-id", "Subject": "Test Message", "From": {"Address": "test@example.com"}}`,
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:           "message not found",
			messageID:      "invalid-id",
			serverResponse: `{"error": "message not found"}`,
			serverStatus:   http.StatusNotFound,
			expectError:    true,
			errorType:      ErrorTypeAPI,
		},
		{
			name:           "empty message ID",
			messageID:      "",
			serverResponse: ``,
			serverStatus:   http.StatusOK,
			expectError:    true,
			errorType:      ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var server *httptest.Server
			if tt.messageID != "" {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Check URL path
					expectedPath := "/api/v1/message/" + tt.messageID
					require.Equal(t, expectedPath, r.URL.Path)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.serverStatus)
					_, _ = w.Write([]byte(tt.serverResponse))
				}))
				defer server.Close()
			}

			var config *Config
			if server != nil {
				config = &Config{
					BaseURL:    server.URL,
					APIPath:    "/api/v1",
					HTTPClient: &http.Client{Timeout: 5 * time.Second},
				}
			} else {
				config = DefaultConfig()
			}

			c, err := NewClient(config)
			require.NoError(t, err)
			defer c.Close()

			result, err := c.GetMessage(t.Context(), tt.messageID)

			if tt.expectError {
				require.Error(t, err)
				var mailpitErr *Error
				require.ErrorAs(t, err, &mailpitErr)
				require.Equal(t, tt.errorType, mailpitErr.Type)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tt.messageID, result.ID)
				require.Equal(t, "Test Message", result.Subject)
			}
		})
	}
}

func TestClient_Close(t *testing.T) {
	t.Parallel()

	client, err := NewClient(nil)
	require.NoError(t, err)

	err = client.Close()
	require.NoError(t, err)
}

func TestClient_DeleteAllMessages(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.True(t, strings.HasSuffix(r.URL.Path, "/messages"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &Config{
		BaseURL:    server.URL,
		APIPath:    "/api/v1",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	c, err := NewClient(config)
	require.NoError(t, err)
	defer c.Close()

	err = c.DeleteAllMessages(t.Context())
	require.NoError(t, err)
}

func TestClient_HealthCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		serverStatus int
		expectError  bool
	}{
		{
			name:         "healthy server",
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "unhealthy server",
			serverStatus: http.StatusInternalServerError,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.True(t, strings.HasSuffix(r.URL.Path, "/info"))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				if tt.serverStatus == http.StatusOK {
					_, _ = w.Write([]byte(`{"version": "test"}`))
				} else {
					_, _ = w.Write([]byte(`{"error": "server error"}`))
				}
			}))
			defer server.Close()

			config := &Config{
				BaseURL:    server.URL,
				APIPath:    "/api/v1",
				HTTPClient: &http.Client{Timeout: 5 * time.Second},
			}

			c, err := NewClient(config)
			require.NoError(t, err)
			defer c.Close()

			err = c.HealthCheck(t.Context())

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test Chaos Operations
func TestClient_GetChaosConfig(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.True(t, strings.HasSuffix(r.URL.Path, "/chaos"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"enabled": true,
			"triggers": {
				"accept_connections": 0.1,
				"reject_senders": 0.2
			}
		}`))
	}))
	defer server.Close()

	config := &Config{
		BaseURL:    server.URL,
		APIPath:    "/api/v1",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	c, err := NewClient(config)
	require.NoError(t, err)
	defer c.Close()

	result, err := c.GetChaosConfig(t.Context())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Enabled)
	require.InEpsilon(t, 0.1, result.Triggers.AcceptConnections, 0.01)
	require.InEpsilon(t, 0.2, result.Triggers.RejectSenders, 0.01)
}

func TestClient_SetChaosConfig(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPut, r.Method)
		require.True(t, strings.HasSuffix(r.URL.Path, "/chaos"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"enabled": true,
			"triggers": {
				"accept_connections": 0.1,
				"reject_senders": 0.2
			}
		}`))
	}))
	defer server.Close()

	config := &Config{
		BaseURL:    server.URL,
		APIPath:    "/api/v1",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	c, err := NewClient(config)
	require.NoError(t, err)
	defer c.Close()

	chaosConfig := &ChaosTriggers{
		AcceptConnections: 0.1,
		RejectSenders:     0.2,
	}

	result, err := c.SetChaosConfig(t.Context(), chaosConfig)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Enabled)
}

// Test Send Operations
func TestClient_SendMessage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.True(t, strings.HasSuffix(r.URL.Path, "/send"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ID": "test-message-id"}`))
	}))
	defer server.Close()

	config := &Config{
		BaseURL:    server.URL,
		APIPath:    "/api/v1",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	c, err := NewClient(config)
	require.NoError(t, err)
	defer c.Close()

	message := &SendMessageRequest{
		From:    Address{Address: "sender@example.com", Name: "Sender"},
		To:      []Address{{Address: "recipient@example.com", Name: "Recipient"}},
		Subject: "Test Message",
		Text:    "This is a test message",
	}

	result, err := c.SendMessage(t.Context(), message)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "test-message-id", result.ID)
}

// Test Message Analysis Operations
func TestClient_GetMessageHeaders(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"Content-Type": ["text/html; charset=utf-8"],
			"Subject": ["Test Message"]
		}`))
	}))
	defer server.Close()

	config := &Config{
		BaseURL:    server.URL,
		APIPath:    "/api/v1",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	c, err := NewClient(config)
	require.NoError(t, err)
	defer c.Close()

	result, err := c.GetMessageHeaders(t.Context(), "test-id")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Contains(t, result, "Content-Type")
	require.Contains(t, result, "Subject")
}

func TestClient_GetMessageHTMLCheck(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Contains(t, r.URL.Path, "/html-check")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"errors": [],
			"warnings": [
				{
					"type": "warning",
					"message": "Missing alt attribute",
					"lastLine": 10,
					"firstColumn": 1,
					"lastColumn": 10,
					"extract": "<img src='test'>",
					"hiliteStart": 0,
					"hiliteLength": 5
				}
			]
		}`))
	}))
	defer server.Close()

	config := &Config{
		BaseURL:    server.URL,
		APIPath:    "/api/v1",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	c, err := NewClient(config)
	require.NoError(t, err)
	defer c.Close()

	result, err := c.GetMessageHTMLCheck(t.Context(), "test-id")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Empty(t, result.Errors)
	require.Len(t, result.Warnings, 1)
	require.Equal(t, "warning", result.Warnings[0].Type)
}
