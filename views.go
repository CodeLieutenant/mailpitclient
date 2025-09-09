package mailpitclient

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

// GetMessageRaw retrieves the raw message source.
func (c *client) GetMessageRaw(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", NewValidationError("message ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/view/%s.raw", id)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &Error{
			Type:    ErrorTypeResponse,
			Message: fmt.Sprintf("failed to read raw content: %v", err),
			Cause:   err,
		}
	}

	return string(body), nil
}

// GetMessagePartHTML retrieves the HTML version of a specific message part.
func (c *client) GetMessagePartHTML(ctx context.Context, messageID, partID string) (string, error) {
	if messageID == "" {
		return "", NewValidationError("message ID cannot be empty")
	}
	if partID == "" {
		return "", NewValidationError("part ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/view/%s/part/%s.html", messageID, partID)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &Error{
			Type:    ErrorTypeResponse,
			Message: fmt.Sprintf("failed to read part HTML content: %v", err),
			Cause:   err,
		}
	}

	return string(body), nil
}

// GetMessagePartText retrieves the text version of a specific message part.
func (c *client) GetMessagePartText(ctx context.Context, messageID, partID string) (string, error) {
	if messageID == "" {
		return "", NewValidationError("message ID cannot be empty")
	}
	if partID == "" {
		return "", NewValidationError("part ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/view/%s/part/%s.text", messageID, partID)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &Error{
			Type:    ErrorTypeResponse,
			Message: fmt.Sprintf("failed to read part text content: %v", err),
			Cause:   err,
		}
	}

	return string(body), nil
}

// GetMessageEvents retrieves events for a specific message.
func (c *client) GetMessageEvents(ctx context.Context, id string) (*EventsResponse, error) {
	if id == "" {
		return nil, NewValidationError("message ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/message/%s/events", id)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var events EventsResponse
	if err = c.parseResponse(resp, &events); err != nil {
		return nil, err
	}

	return &events, nil
}
