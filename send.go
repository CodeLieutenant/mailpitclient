package mailpitclient

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// SendMessage sends a message via the HTTP API.
func (c *client) SendMessage(ctx context.Context, message *SendMessageRequest) (*SendMessageResponse, error) {
	if message == nil {
		return nil, NewValidationError("message cannot be nil")
	}

	endpoint := "/send"

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(message); err != nil {
		return nil, &Error{
			Type:    ErrorTypeRequest,
			Message: "failed to encode send message request",
			Cause:   err,
		}
	}

	resp, err := c.makeRequest(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return nil, err
	}

	var result SendMessageResponse
	if err = c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
