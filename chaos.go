package mailpit_go_api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// GetChaosConfig retrieves the current chaos triggers configuration.
// This API route will return an error if Chaos is not enabled at runtime.
func (c *client) GetChaosConfig(ctx context.Context) (*ChaosResponse, error) {
	endpoint := "/chaos"

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var chaos ChaosResponse
	if err = c.parseResponse(resp, &chaos); err != nil {
		return nil, err
	}

	return &chaos, nil
}

// SetChaosConfig sets the chaos triggers configuration and returns the updated values.
// This API route will return an error if Chaos is not enabled at runtime.
// If any triggers are omitted from the request, then those are reset to their
// default values with a 0% probability (ie: disabled).
// Setting a blank config will reset all triggers to their default values.
func (c *client) SetChaosConfig(ctx context.Context, config *ChaosTriggers) (*ChaosResponse, error) {
	endpoint := "/chaos"

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(config); err != nil {
		return nil, &Error{
			Type:    ErrorTypeRequest,
			Message: "failed to encode chaos config",
			Cause:   err,
		}
	}

	resp, err := c.makeRequest(ctx, http.MethodPut, endpoint, &body)
	if err != nil {
		return nil, err
	}

	var chaos ChaosResponse
	if err = c.parseResponse(resp, &chaos); err != nil {
		return nil, err
	}

	return &chaos, nil
}
