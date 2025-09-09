package mailpitclient

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// SetTags sets the list of available tags on the server.
func (c *client) SetTags(ctx context.Context, tags []string) ([]string, error) {
	endpoint := "/tags"

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(tags); err != nil {
		return nil, &Error{
			Type:    ErrorTypeRequest,
			Message: "failed to encode tags",
			Cause:   err,
		}
	}

	resp, err := c.makeRequest(ctx, http.MethodPut, endpoint, &body)
	if err != nil {
		return nil, err
	}

	var result []string
	if err = c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// SetMessageTags sets tags for specific messages.
func (c *client) SetMessageTags(ctx context.Context, tag string, messageIDs []string) error {
	if tag == "" {
		return NewValidationError("tag cannot be empty")
	}
	if len(messageIDs) == 0 {
		return NewValidationError("message IDs cannot be empty")
	}

	endpoint := "/tags/" + tag

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(messageIDs); err != nil {
		return &Error{
			Type:    ErrorTypeRequest,
			Message: "failed to encode message IDs",
			Cause:   err,
		}
	}

	resp, err := c.makeRequest(ctx, http.MethodPut, endpoint, &body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// DeleteTag deletes a tag from the server.
func (c *client) DeleteTag(ctx context.Context, tag string) error {
	if tag == "" {
		return NewValidationError("tag cannot be empty")
	}

	endpoint := "/tags/" + tag

	resp, err := c.makeRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
