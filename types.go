package mailpitclient

import (
	"encoding/json"
	"net/url"
	"strconv"
	"time"
)

// AttachmentList handles both array and number formats from mailpit API
type AttachmentList []Attachment

// UnmarshalJSON handles the case where mailpit returns 0 instead of empty array
func (al *AttachmentList) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as array first
	var attachments []Attachment
	if err := json.Unmarshal(data, &attachments); err == nil {
		*al = attachments

		return nil
	}

	// If that fails, try as number (mailpit returns 0 for no attachments)
	var num int
	if err := json.Unmarshal(data, &num); err == nil {
		*al = AttachmentList{}

		return nil
	}

	// If both fail, return empty list
	*al = AttachmentList{}

	return nil
}

// Message represents an email message in Mailpit.
type Message struct {
	Date        time.Time      `json:"Date"`
	Created     time.Time      `json:"Created"`
	From        Address        `json:"From"`
	MessageID   string         `json:"MessageID"`
	ID          string         `json:"ID"`
	HTML        string         `json:"HTML"`
	Text        string         `json:"Text"`
	Subject     string         `json:"Subject"`
	Cc          []Address      `json:"Cc,omitempty"`
	ReplyTo     []Address      `json:"ReplyTo,omitempty"`
	Bcc         []Address      `json:"Bcc,omitempty"`
	Inline      AttachmentList `json:"Inline,omitempty"`
	Attachments AttachmentList `json:"Attachments,omitempty"`
	Tags        []string       `json:"Tags,omitempty"`
	To          []Address      `json:"To"`
	Size        int            `json:"Size"`
	Read        bool           `json:"Read"`
}

// Address represents an email address with optional name.
type Address struct {
	Address string `json:"Address"`
	Name    string `json:"Name"`
}

// Attachment represents an email attachment.
type Attachment struct {
	PartID      string `json:"PartID"`
	FileName    string `json:"FileName"`
	ContentType string `json:"ContentType"`
	Size        int    `json:"Size"`
}

// MessagesResponse represents the response from the messages API.
type MessagesResponse struct {
	Tags          []string  `json:"tags"`
	Messages      []Message `json:"messages"`
	Total         int       `json:"total"`
	Unread        int       `json:"unread"`
	Count         int       `json:"count"`
	Start         int       `json:"start"`
	MessagesCount int       `json:"messages_count"`
}

// ServerInfo represents server information and status.
type ServerInfo struct {
	Tags          any          `json:"Tags"`
	Version       string       `json:"Version"`
	LatestVersion string       `json:"LatestVersion,omitempty"`
	Database      string       `json:"Database"`
	RuntimeStats  RuntimeStats `json:"RuntimeStats"`
	DatabaseSize  int64        `json:"DatabaseSize"`
	Messages      int          `json:"Messages"`
	Unread        int          `json:"Unread"`
}

// RuntimeStats represents runtime statistics
type RuntimeStats struct {
	Uptime           int64 `json:"Uptime"`
	Memory           int64 `json:"Memory"`
	MessagesDeleted  int   `json:"MessagesDeleted"`
	SMTPAccepted     int   `json:"SMTPAccepted"`
	SMTPAcceptedSize int64 `json:"SMTPAcceptedSize"`
	SMTPRejected     int   `json:"SMTPRejected"`
	SMTPIgnored      int   `json:"SMTPIgnored"`
}

// Stats represents server statistics.
type Stats struct {
	CreatedAt string   `json:"created_at"`
	Tags      []string `json:"tags"`
	Total     int      `json:"total"`
	Unread    int      `json:"unread"`
}

// ListOptions represents options for listing messages.
type ListOptions struct {
	Query string `json:"query,omitempty"`
	Tag   string `json:"tag,omitempty"`
	Sort  string `json:"sort,omitempty"`
	Start int    `json:"start,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

// SearchOptions represents options for searching messages.
type SearchOptions struct {
	Tag   string `json:"tag,omitempty"`
	Sort  string `json:"sort,omitempty"`
	Start int    `json:"start,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

// ToURLValues converts ListOptions to url.Values for query parameters.
func (opts *ListOptions) ToURLValues() url.Values {
	values := url.Values{}

	if opts == nil {
		return values
	}

	if opts.Start > 0 {
		values.Set("start", strconv.Itoa(opts.Start))
	}

	if opts.Limit > 0 {
		values.Set("limit", strconv.Itoa(opts.Limit))
	}

	if opts.Query != "" {
		values.Set("query", opts.Query)
	}

	if opts.Tag != "" {
		values.Set("tag", opts.Tag)
	}

	if opts.Sort != "" {
		values.Set("sort", opts.Sort)
	}

	return values
}

// ToURLValues converts SearchOptions to url.Values for query parameters.
func (opts *SearchOptions) ToURLValues() url.Values {
	values := url.Values{}

	if opts == nil {
		return values
	}

	if opts.Start > 0 {
		values.Set("start", strconv.Itoa(opts.Start))
	}

	if opts.Limit > 0 {
		values.Set("limit", strconv.Itoa(opts.Limit))
	}

	if opts.Tag != "" {
		values.Set("tag", opts.Tag)
	}

	if opts.Sort != "" {
		values.Set("sort", opts.Sort)
	}

	return values
}

// MessageSummary represents a lightweight version of a message for listings.
type MessageSummary struct {
	Date    time.Time `json:"Date"`
	Created time.Time `json:"Created"`
	From    Address   `json:"From"`
	ID      string    `json:"ID"`
	Subject string    `json:"Subject"`
	To      []Address `json:"To"`
	Tags    []string  `json:"Tags,omitempty"`
	Size    int       `json:"Size"`
	Read    bool      `json:"Read"`
}

// HTMLCheckResponse represents response from HTML check endpoint.
type HTMLCheckResponse struct {
	Errors   []HTMLCheckError `json:"errors,omitempty"`
	Warnings []HTMLCheckError `json:"warnings,omitempty"`
}

// HTMLCheckError represents an HTML validation error or warning.
type HTMLCheckError struct {
	Type         string `json:"type"`
	Message      string `json:"message"`
	Extract      string `json:"extract"`
	LastLine     int    `json:"lastLine"`
	FirstColumn  int    `json:"firstColumn"`
	LastColumn   int    `json:"lastColumn"`
	HiliteStart  int    `json:"hiliteStart"`
	HiliteLength int    `json:"hiliteLength"`
}

// LinkCheckResponse represents response from link check endpoint.
type LinkCheckResponse struct {
	Links []LinkCheck `json:"links,omitempty"`
}

// LinkCheck represents a checked link.
type LinkCheck struct {
	Status any    `json:"status"`
	URL    string `json:"url"`
	Error  string `json:"error,omitempty"`
}

// SpamAssassinCheckResponse represents response from SpamAssassin check endpoint.
type SpamAssassinCheckResponse struct {
	Symbols []SpamAssassinSymbol `json:"symbols,omitempty"`
	Report  []SpamAssassinReport `json:"report,omitempty"`
	Score   float64              `json:"score"`
}

// SpamAssassinSymbol represents a SpamAssassin symbol.
type SpamAssassinSymbol struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Score       float64 `json:"score"`
}

// SpamAssassinReport represents a SpamAssassin report line.
type SpamAssassinReport struct {
	Description string  `json:"description"`
	Score       float64 `json:"score"`
}

// ReleaseMessageRequest represents a request to release a message.
type ReleaseMessageRequest struct {
	Host string   `json:"host,omitempty"`
	To   []string `json:"to"`
	Port int      `json:"port,omitempty"`
}

// SendMessageRequest represents a request to send a message via HTTP.
type SendMessageRequest struct {
	Headers     map[string]string `json:"headers,omitempty"`
	From        Address           `json:"from"`
	Subject     string            `json:"subject"`
	Text        string            `json:"text,omitempty"`
	HTML        string            `json:"html,omitempty"`
	To          []Address         `json:"to"`
	Cc          []Address         `json:"cc,omitempty"`
	Bcc         []Address         `json:"bcc,omitempty"`
	ReplyTo     []Address         `json:"reply-to,omitempty"`
	Attachments []SendAttachment  `json:"attachments,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
}

// SendAttachment represents an attachment for sending.
type SendAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content-type,omitempty"`
	Content     string `json:"content"` // base64 encoded
}

// SendMessageResponse represents response from send message endpoint.
type SendMessageResponse struct {
	ID string `json:"ID"`
}

// WebUIConfig represents web UI configuration.
type WebUIConfig struct {
	Label               string       `json:"Label"`
	MessageRelay        MessageRelay `json:"MessageRelay"`
	SpamAssassin        bool         `json:"SpamAssassin"`
	ChaosEnabled        bool         `json:"ChaosEnabled"`
	DuplicatesIgnored   bool         `json:"DuplicatesIgnored"`
	HideDeleteAllButton bool         `json:"HideDeleteAllButton"`
}

// MessageRelay represents message relay configuration
type MessageRelay struct {
	SMTPServer         string `json:"SMTPServer"`
	ReturnPath         string `json:"ReturnPath"`
	AllowedRecipients  string `json:"AllowedRecipients"`
	BlockedRecipients  string `json:"BlockedRecipients"`
	OverrideFrom       string `json:"OverrideFrom"`
	RecipientAllowlist string `json:"RecipientAllowlist"`
	Enabled            bool   `json:"Enabled"`
	PreserveMessageIDs bool   `json:"PreserveMessageIDs"`
}

// ChaosResponse represents response from chaos endpoints.
type ChaosResponse struct {
	Enabled  bool          `json:"enabled"`
	Triggers ChaosTriggers `json:"triggers"`
}

// ChaosTriggers represents chaos testing triggers configuration.
type ChaosTriggers struct {
	AcceptConnections float64 `json:"accept_connections,omitempty"`
	RejectSenders     float64 `json:"reject_senders,omitempty"`
	RejectRecipients  float64 `json:"reject_recipients,omitempty"`
	RejectAuth        float64 `json:"reject_auth,omitempty"`
	RejectData        float64 `json:"reject_data,omitempty"`
	DelayConnections  float64 `json:"delay_connections,omitempty"`
	DelayAuth         float64 `json:"delay_auth,omitempty"`
	DelayMailFrom     float64 `json:"delay_mail_from,omitempty"`
	DelayRcptTo       float64 `json:"delay_rcpt_to,omitempty"`
	DelayData         float64 `json:"delay_data,omitempty"`
}

// EventsResponse represents response from message events endpoint.
type EventsResponse struct {
	Events []MessageEvent `json:"events"`
}

// MessageEvent represents a message event.
type MessageEvent struct {
	Timestamp time.Time `json:"Timestamp"`
	Data      any       `json:"Data,omitempty"`
	ID        string    `json:"ID"`
	Type      string    `json:"Type"`
}
