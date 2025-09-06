// package mailpit_go_api provides a production-ready client for interacting with Mailpit API.
// Mailpit is a popular email testing tool that provides a REST API for managing emails.
package mailpit_go_api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client represents a Mailpit API client.
type Client interface {
	// Message operations
	ListMessages(ctx context.Context, opts *ListOptions) (*MessagesResponse, error)
	GetMessage(ctx context.Context, id string) (*Message, error)
	GetMessageSource(ctx context.Context, id string) (string, error)
	GetMessageHeaders(ctx context.Context, id string) (map[string][]string, error)
	GetMessageHTMLCheck(ctx context.Context, id string) (*HTMLCheckResponse, error)
	GetMessageLinkCheck(ctx context.Context, id string) (*LinkCheckResponse, error)
	GetMessageSpamAssassinCheck(ctx context.Context, id string) (*SpamAssassinCheckResponse, error)
	GetMessagePart(ctx context.Context, messageID, partID string) ([]byte, error)
	GetMessagePartThumbnail(ctx context.Context, messageID, partID string) ([]byte, error)
	GetMessageAttachment(ctx context.Context, messageID, attachmentID string) ([]byte, error)
	DeleteMessage(ctx context.Context, id string) error
	DeleteAllMessages(ctx context.Context) error
	MarkMessageRead(ctx context.Context, id string) error
	MarkMessageUnread(ctx context.Context, id string) error
	ReleaseMessage(ctx context.Context, id string, releaseData *ReleaseMessageRequest) error
	SearchMessages(ctx context.Context, query string, opts *SearchOptions) (*MessagesResponse, error)
	DeleteSearchResults(ctx context.Context, query string) error

	// Send operations
	SendMessage(ctx context.Context, message *SendMessageRequest) (*SendMessageResponse, error)

	// Tags operations
	GetTags(ctx context.Context) ([]string, error)
	SetTags(ctx context.Context, tags []string) ([]string, error)
	SetMessageTags(ctx context.Context, tag string, messageIDs []string) error
	DeleteTag(ctx context.Context, tag string) error

	// View operations
	GetMessageHTML(ctx context.Context, id string) (string, error)
	GetMessageText(ctx context.Context, id string) (string, error)

	// Server operations
	GetServerInfo(ctx context.Context) (*ServerInfo, error)
	GetWebUIConfig(ctx context.Context) (*WebUIConfig, error)
	HealthCheck(ctx context.Context) error
	Ping(ctx context.Context) error

	// Statistics
	GetStats(ctx context.Context) (*Stats, error)

	// Chaos testing operations
	GetChaosConfig(ctx context.Context) (*ChaosResponse, error)
	SetChaosConfig(ctx context.Context, config *ChaosTriggers) (*ChaosResponse, error)

	// Utility methods
	Close() error
}

// Config holds the configuration for the Mailpit client.
type Config struct {
	HTTPClient *http.Client
	BaseURL    string
	APIPath    string
	Username   string
	Password   string
	APIKey     string
	UserAgent  string
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		BaseURL:    "http://localhost:8025",
		APIPath:    "/api/v1",
		Timeout:    30 * time.Second,
		UserAgent:  "mailpit-go-client/1.0.0",
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// client is the concrete implementation of the Client interface.
type client struct {
	config    *Config
	baseURL   *url.URL
	apiURL    string
	userAgent string
}

// NewClient creates a new Mailpit client with the given configuration.
func NewClient(config *Config) (Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Validate configuration
	if config.BaseURL == "" {
		return nil, &Error{
			Type:    ErrorTypeConfig,
			Message: "BaseURL is required",
		}
	}

	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, &Error{
			Type:    ErrorTypeConfig,
			Message: fmt.Sprintf("invalid BaseURL: %v", err),
			Cause:   err,
		}
	}

	if config.APIPath == "" {
		config.APIPath = "/api/v1"
	}

	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: config.Timeout,
		}
	}

	if config.UserAgent == "" {
		config.UserAgent = "mailpit-go-client/1.0.0"
	}

	apiURL := baseURL.String() + config.APIPath

	return &client{
		config:    config,
		baseURL:   baseURL,
		apiURL:    apiURL,
		userAgent: config.UserAgent,
	}, nil
}

// Close closes the client and releases any resources.
func (c *client) Close() error {
	// Close HTTP client if we own it
	if transport, ok := c.config.HTTPClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}

	return nil
}

// makeRequest performs an HTTP request with proper error handling and retries.
//
//nolint:unparam
func (c *client) makeRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	u := c.apiURL + endpoint

	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.config.RetryDelay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, u, body)
		if err != nil {
			return nil, &Error{
				Type:    ErrorTypeRequest,
				Message: fmt.Sprintf("failed to create request: %v", err),
				Cause:   err,
			}
		}

		// Set headers
		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// Add authentication if configured
		if c.config.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
		} else if c.config.Username != "" && c.config.Password != "" {
			req.SetBasicAuth(c.config.Username, c.config.Password)
		}

		resp, err := c.config.HTTPClient.Do(req)
		if err != nil {
			lastErr = &Error{
				Type:    ErrorTypeNetwork,
				Message: fmt.Sprintf("request failed: %v", err),
				Cause:   err,
			}

			continue
		}

		// Check for successful response
		if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
			return resp, nil
		}

		// Handle HTTP errors
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)

		lastErr = &Error{
			Type:       ErrorTypeAPI,
			Message:    fmt.Sprintf("API request failed with status %d", resp.StatusCode),
			StatusCode: resp.StatusCode,
			Response:   string(b),
		}

		// Don't retry on 4xx errors (except rate limiting)
		if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode < http.StatusInternalServerError && resp.StatusCode != http.StatusTooManyRequests {
			break
		}
	}

	return nil, lastErr
}

// parseResponse parses a JSON response into the given struct.
func (c *client) parseResponse(resp *http.Response, target any) error {
	defer resp.Body.Close()

	if target == nil {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Error{
			Type:    ErrorTypeResponse,
			Message: fmt.Sprintf("failed to read response body: %v", err),
			Cause:   err,
		}
	}

	if len(body) == 0 {
		return nil
	}

	if err = json.Unmarshal(body, target); err != nil {
		return &Error{
			Type:    ErrorTypeResponse,
			Message: fmt.Sprintf("failed to parse JSON response: %v", err),
			Cause:   err,
		}
	}

	return nil
}
