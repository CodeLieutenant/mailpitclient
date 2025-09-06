package mailpit_go_api

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// GetMessageHTML retrieves the HTML view of a specific message.
func (c *client) GetMessageHTML(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", NewValidationError("message ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/view/%s.html", id)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &Error{
			Type:    ErrorTypeResponse,
			Message: fmt.Sprintf("failed to read HTML content: %v", err),
			Cause:   err,
		}
	}

	return string(body), nil
}

// GetMessageText retrieves the text view of a specific message.
func (c *client) GetMessageText(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", NewValidationError("message ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/view/%s.txt", id)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &Error{
			Type:    ErrorTypeResponse,
			Message: fmt.Sprintf("failed to read text content: %v", err),
			Cause:   err,
		}
	}

	return string(body), nil
}
