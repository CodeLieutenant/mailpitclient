package mailpitclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ListMessages retrieves a list of messages with optional filtering and pagination.
func (c *client) ListMessages(ctx context.Context, opts *ListOptions) (*MessagesResponse, error) {
	endpoint := "/messages"

	if opts != nil {
		params := opts.ToURLValues()
		if len(params) > 0 {
			endpoint += "?" + params.Encode()
		}
	}

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result MessagesResponse
	if err = c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetMessage retrieves a specific message by its ID.
func (c *client) GetMessage(ctx context.Context, id string) (*Message, error) {
	if id == "" {
		return nil, NewValidationError("message ID cannot be empty")
	}

	resp, err := c.makeRequest(ctx, http.MethodGet, "/message/"+id, nil)
	if err != nil {
		return nil, err
	}

	var message Message
	if err = c.parseResponse(resp, &message); err != nil {
		return nil, err
	}

	return &message, nil
}

// GetMessageSource retrieves the raw source of a message.
func (c *client) GetMessageSource(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", NewValidationError("message ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/messages/%s/source", id)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &Error{
			Type:    ErrorTypeResponse,
			Message: fmt.Sprintf("failed to read message source: %v", err),
			Cause:   err,
		}
	}

	return string(body), nil
}

// DeleteMessage deletes a specific message by its ID.
func (c *client) DeleteMessage(ctx context.Context, id string) error {
	if id == "" {
		return NewValidationError("message ID cannot be empty")
	}

	resp, err := c.makeRequest(ctx, http.MethodDelete, "/messages/"+id, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// DeleteAllMessages deletes all messages from the mailbox.
func (c *client) DeleteAllMessages(ctx context.Context) error {
	endpoint := "/messages"

	resp, err := c.makeRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// SearchMessages searches for messages matching the given query.
func (c *client) SearchMessages(ctx context.Context, query string, opts *SearchOptions) (*MessagesResponse, error) {
	if query == "" {
		return nil, NewValidationError("search query cannot be empty")
	}

	endpoint := "/search"
	params := []string{"query=" + url.QueryEscape(query)}

	if opts != nil {
		urlValues := opts.ToURLValues()
		for key, values := range urlValues {
			for _, value := range values {
				params = append(params, key+"="+value)
			}
		}
	}

	if len(params) > 0 {
		endpoint += "?" + strings.Join(params, "&")
	}

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result MessagesResponse
	if err = c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetMessageAttachment retrieves a specific attachment from a message.
func (c *client) GetMessageAttachment(ctx context.Context, messageID, attachmentID string) ([]byte, error) {
	if messageID == "" {
		return nil, NewValidationError("message ID cannot be empty")
	}
	if attachmentID == "" {
		return nil, NewValidationError("attachment ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/messages/%s/part/%s", messageID, attachmentID)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &Error{
			Type:    ErrorTypeResponse,
			Message: fmt.Sprintf("failed to read attachment data: %v", err),
			Cause:   err,
		}
	}

	return data, nil
}

// MarkMessageRead marks a message as read.
func (c *client) MarkMessageRead(ctx context.Context, id string) error {
	if id == "" {
		return NewValidationError("message ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/messages/%s/read", id)

	resp, err := c.makeRequest(ctx, http.MethodPut, endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// MarkMessageUnread marks a message as unread.
func (c *client) MarkMessageUnread(ctx context.Context, id string) error {
	if id == "" {
		return NewValidationError("message ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/messages/%s/unread", id)

	resp, err := c.makeRequest(ctx, http.MethodPut, endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetMessageHeaders retrieves the headers of a specific message.
func (c *client) GetMessageHeaders(ctx context.Context, id string) (map[string][]string, error) {
	if id == "" {
		return nil, NewValidationError("message ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/message/%s/headers", id)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var headers map[string][]string
	if err = c.parseResponse(resp, &headers); err != nil {
		return nil, err
	}

	return headers, nil
}

// GetMessageHTMLCheck performs HTML validation on a message.
func (c *client) GetMessageHTMLCheck(ctx context.Context, id string) (*HTMLCheckResponse, error) {
	if id == "" {
		return nil, NewValidationError("message ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/message/%s/html-check", id)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result HTMLCheckResponse
	if err = c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetMessageLinkCheck performs link checking on a message.
func (c *client) GetMessageLinkCheck(ctx context.Context, id string) (*LinkCheckResponse, error) {
	if id == "" {
		return nil, NewValidationError("message ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/message/%s/link-check", id)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result LinkCheckResponse
	if err = c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetMessageSpamAssassinCheck performs SpamAssassin checking on a message.
func (c *client) GetMessageSpamAssassinCheck(ctx context.Context, id string) (*SpamAssassinCheckResponse, error) {
	if id == "" {
		return nil, NewValidationError("message ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/message/%s/sa-check", id)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result SpamAssassinCheckResponse
	if err = c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetMessagePart retrieves a specific part of a message.
func (c *client) GetMessagePart(ctx context.Context, messageID, partID string) ([]byte, error) {
	if messageID == "" {
		return nil, NewValidationError("message ID cannot be empty")
	}
	if partID == "" {
		return nil, NewValidationError("part ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/message/%s/part/%s", messageID, partID)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &Error{
			Type:    ErrorTypeResponse,
			Message: fmt.Sprintf("failed to read message part data: %v", err),
			Cause:   err,
		}
	}

	return data, nil
}

// GetMessagePartThumbnail retrieves a thumbnail for a specific message part.
func (c *client) GetMessagePartThumbnail(ctx context.Context, messageID, partID string) ([]byte, error) {
	if messageID == "" {
		return nil, NewValidationError("message ID cannot be empty")
	}
	if partID == "" {
		return nil, NewValidationError("part ID cannot be empty")
	}

	endpoint := fmt.Sprintf("/message/%s/part/%s/thumb", messageID, partID)

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &Error{
			Type:    ErrorTypeResponse,
			Message: fmt.Sprintf("failed to read thumbnail data: %v", err),
			Cause:   err,
		}
	}

	return data, nil
}

// ReleaseMessage releases a message via SMTP relay.
func (c *client) ReleaseMessage(ctx context.Context, id string, releaseData *ReleaseMessageRequest) error {
	if id == "" {
		return NewValidationError("message ID cannot be empty")
	}
	if releaseData == nil {
		return NewValidationError("release data cannot be nil")
	}

	endpoint := fmt.Sprintf("/message/%s/release", id)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(releaseData); err != nil {
		return &Error{
			Type:    ErrorTypeRequest,
			Message: "failed to encode release message request",
			Cause:   err,
		}
	}

	resp, err := c.makeRequest(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// DeleteSearchResults deletes messages matching the given search query.
func (c *client) DeleteSearchResults(ctx context.Context, query string) error {
	if query == "" {
		return NewValidationError("search query cannot be empty")
	}

	endpoint := "/search?query=" + url.QueryEscape(query)

	resp, err := c.makeRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
