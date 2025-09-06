// nolint:goconst
package mailpit_go_api

import (
	"context"
	"net/http"
)

// GetServerInfo retrieves server information and configuration.
func (c *client) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	endpoint := "/info"

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var info ServerInfo
	if err = c.parseResponse(resp, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// HealthCheck performs a health check against the server.
// This is a simple check that verifies the server is responding.
func (c *client) HealthCheck(ctx context.Context) error {
	endpoint := "/info"

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// If we got here, the server is healthy
	return nil
}

// GetStats retrieves server statistics including message counts and tags.
func (c *client) GetStats(ctx context.Context) (*Stats, error) {
	endpoint := "/stats"

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var stats Stats
	if err = c.parseResponse(resp, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetTags retrieves all available message tags from the server.
func (c *client) GetTags(ctx context.Context) ([]string, error) {
	endpoint := "/tags"

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var tags []string
	if err = c.parseResponse(resp, &tags); err != nil {
		return nil, err
	}

	return tags, nil
}

// Ping performs a simple ping to check if the server is reachable.
// This is a lightweight alternative to HealthCheck.
func (c *client) Ping(ctx context.Context) error {
	endpoint := "/info"

	resp, err := c.makeRequest(ctx, http.MethodHead, endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetWebUIConfig retrieves the web UI configuration.
func (c *client) GetWebUIConfig(ctx context.Context) (*WebUIConfig, error) {
	endpoint := "/webui"

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var config WebUIConfig
	if err = c.parseResponse(resp, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
