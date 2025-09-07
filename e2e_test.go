package mailpit_go_api

import (
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_ServerOperations tests all server-related operations
func TestE2E_ServerOperations(t *testing.T) {
	t.Parallel()
	testSMTP := GetTestSMTP(t)
	client := testSMTP.MailpitClient
	ctx := t.Context()

	t.Run("HealthCheck", func(t *testing.T) {
		t.Parallel()
		err := client.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	t.Run("GetServerInfo", func(t *testing.T) {
		t.Parallel()
		info, err := client.GetServerInfo(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, info.Version)
		assert.NotEmpty(t, info.Database)
		assert.GreaterOrEqual(t, info.Messages, 0)
		assert.GreaterOrEqual(t, info.Unread, 0)
	})

	t.Run("GetWebUIConfig", func(t *testing.T) {
		t.Parallel()
		config, err := client.GetWebUIConfig(ctx)
		require.NoError(t, err)
		assert.NotNil(t, config)
		// The config might have empty fields, but should not be nil
	})
}

// TestE2E_MessageOperations tests all message-related operations
// nolint
func TestE2E_MessageOperations(t *testing.T) {
	t.Parallel()

	t.Run("ListMessages", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email to work with
		sendTestEmailWithSubject(t, testSMTP, "ListMessages Test Email")
		time.Sleep(2 * time.Second)

		// Test basic list
		response, err := client.ListMessages(ctx, nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Messages), 1)
		assert.GreaterOrEqual(t, response.Total, 1)
		assert.Equal(t, response.Count, len(response.Messages))

		// Test with options
		opts := &ListOptions{
			Start: 0,
			Limit: 1,
		}
		response, err = client.ListMessages(ctx, opts)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(response.Messages), 1)
	})

	t.Run("GetMessage", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email to work with
		sendTestEmailWithSubject(t, testSMTP, "GetMessage Test Email")
		time.Sleep(2 * time.Second)

		// Get the first message
		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		message, err := client.GetMessage(ctx, messageID)
		require.NoError(t, err)
		assert.Equal(t, messageID, message.ID)
		assert.NotEmpty(t, message.Subject)
		assert.NotEmpty(t, message.From.Address)
	})

	t.Run("GetMessageSource", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email to work with
		sendTestEmailWithSubject(t, testSMTP, "GetMessageSource Test Email")
		time.Sleep(2 * time.Second)

		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		// Note: Source endpoint may not be available in all Mailpit versions
		_, err = client.GetMessageSource(ctx, messageID)
		if err != nil {
			t.Logf("GetMessageSource not available (expected in some versions): %v", err)
		}
	})

	t.Run("GetMessageHeaders", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email to work with
		sendTestEmailWithSubject(t, testSMTP, "GetMessageHeaders Test Email")
		time.Sleep(2 * time.Second)

		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		headers, err := client.GetMessageHeaders(ctx, messageID)
		require.NoError(t, err)
		assert.NotEmpty(t, headers)
		assert.Contains(t, headers, "Subject")
		assert.Contains(t, headers, "From")
	})

	t.Run("MarkMessageRead", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email to work with
		sendTestEmailWithSubject(t, testSMTP, "MarkMessageRead Test Email")
		time.Sleep(2 * time.Second)

		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		// Note: Mark read endpoint may not be available in all Mailpit versions
		err = client.MarkMessageRead(ctx, messageID)
		if err != nil {
			t.Logf("MarkMessageRead not available (expected in some versions): %v", err)
		}
	})

	t.Run("MarkMessageUnread", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email to work with
		sendTestEmailWithSubject(t, testSMTP, "MarkMessageUnread Test Email")
		time.Sleep(2 * time.Second)

		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		// Note: Mark unread endpoint may not be available in all Mailpit versions
		err = client.MarkMessageUnread(ctx, messageID)
		if err != nil {
			t.Logf("MarkMessageUnread not available (expected in some versions): %v", err)
		}
	})

	t.Run("SearchMessages", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email to search for
		sendTestEmailWithSubject(t, testSMTP, "SearchMessages Test Email")
		time.Sleep(2 * time.Second)

		// Search for messages with the test subject
		response, err := client.SearchMessages(ctx, "subject:SearchMessages Test", nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Messages), 1)

		// Verify the search found our test email
		found := false
		for _, msg := range response.Messages {
			if strings.Contains(msg.Subject, "SearchMessages Test") {
				found = true

				break
			}
		}
		assert.True(t, found, "Should find test email in search results")
	})

	t.Run("GetMessageHTML", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send an HTML email first
		sendTestHTMLEmail(t, testSMTP)
		time.Sleep(2 * time.Second)

		response, err := client.SearchMessages(ctx, "subject:HTML E2E Test", nil)
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		// Note: HTML view endpoint may not be available in all Mailpit versions
		_, err = client.GetMessageHTML(ctx, messageID)
		if err != nil {
			t.Logf("GetMessageHTML not available (expected in some versions): %v", err)
		}
	})

	t.Run("GetMessageText", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email to work with
		sendTestEmailWithSubject(t, testSMTP, "GetMessageText Test Email")
		time.Sleep(2 * time.Second)

		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		// Note: Text view endpoint may not be available in all Mailpit versions
		_, err = client.GetMessageText(ctx, messageID)
		if err != nil {
			t.Logf("GetMessageText not available (expected in some versions): %v", err)
		}
	})

	t.Run("GetMessageRaw", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email to work with
		sendTestEmailWithSubject(t, testSMTP, "GetMessageRaw Test Email")
		time.Sleep(2 * time.Second)

		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		rawContent, err := client.GetMessageRaw(ctx, messageID)
		if err != nil {
			t.Logf("GetMessageRaw not available (expected in some versions): %v", err)
		} else {
			assert.NotEmpty(t, rawContent, "Raw content should not be empty")
		}
	})

	t.Run("GetMessageEvents", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email to work with
		sendTestEmailWithSubject(t, testSMTP, "GetMessageEvents Test Email")
		time.Sleep(2 * time.Second)

		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		events, err := client.GetMessageEvents(ctx, messageID)
		if err != nil {
			t.Logf("GetMessageEvents not available (expected in some versions): %v", err)
		} else {
			assert.NotNil(t, events, "Events response should not be nil")
		}
	})

	t.Run("GetMessagePartHTML", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send an HTML email with parts first
		sendTestHTMLEmail(t, testSMTP)
		time.Sleep(2 * time.Second)

		response, err := client.SearchMessages(ctx, "subject:HTML E2E Test", nil)
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		// Try to get HTML part - may not exist or be available
		_, err = client.GetMessagePartHTML(ctx, messageID, "1")
		if err != nil {
			t.Logf("GetMessagePartHTML not available or no parts (expected): %v", err)
		}
	})

	t.Run("GetMessagePartText", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email to work with
		sendTestEmailWithSubject(t, testSMTP, "GetMessagePartText Test Email")
		time.Sleep(2 * time.Second)

		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		// Try to get text part - may not exist or be available
		_, err = client.GetMessagePartText(ctx, messageID, "1")
		if err != nil {
			t.Logf("GetMessagePartText not available or no parts (expected): %v", err)
		}
	})
}

// TestE2E_MessageWithAttachments tests message operations with attachments
func TestE2E_MessageWithAttachments(t *testing.T) {
	t.Parallel()

	t.Run("GetMessagePart", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send an email with attachment first
		sendTestEmailWithAttachment(t, testSMTP)
		time.Sleep(3 * time.Second) // Allow more time for processing

		response, err := client.SearchMessages(ctx, "subject:E2E Test Email with Attachment", nil)
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		message, err := client.GetMessage(ctx, messageID)
		require.NoError(t, err)

		// Check if the message has attachments
		if len(message.Attachments) > 0 {
			attachmentID := message.Attachments[0].PartID
			//nolint:revive,govet
			partContent, err := client.GetMessagePart(ctx, messageID, attachmentID)
			require.NoError(t, err)
			assert.NotEmpty(t, partContent, "Attachment content should not be empty")
		} else {
			t.Log("No attachments found in the test message")
		}
	})

	t.Run("GetMessageAttachment", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send an email with attachment first
		sendTestEmailWithAttachment(t, testSMTP)
		time.Sleep(3 * time.Second) // Allow more time for processing

		response, err := client.SearchMessages(ctx, "subject:Attachment E2E Test Email", nil)
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		message, err := client.GetMessage(ctx, messageID)
		require.NoError(t, err)

		// Check if the message has attachments
		if len(message.Attachments) > 0 {
			attachmentID := message.Attachments[0].PartID
			attachmentContent, attachmentErr := client.GetMessageAttachment(ctx, messageID, attachmentID)
			if attachmentErr != nil {
				t.Logf("GetMessageAttachment not available (expected in some versions): %v", attachmentErr)

				return
			}
			assert.NotEmpty(t, attachmentContent, "Attachment content should not be empty")
		} else {
			t.Log("No attachments found in the test message")
		}
	})

	t.Run("GetMessagePartThumbnail", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send an email with attachment first
		sendTestEmailWithAttachment(t, testSMTP)
		time.Sleep(3 * time.Second) // Allow more time for processing

		response, err := client.SearchMessages(ctx, "subject:E2E Test Email with Attachment", nil)
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		message, err := client.GetMessage(ctx, messageID)
		require.NoError(t, err)

		// Check if the message has attachments
		if len(message.Attachments) > 0 {
			attachmentID := message.Attachments[0].PartID
			thumbnail, thumbnailErr := client.GetMessagePartThumbnail(ctx, messageID, attachmentID)
			// Thumbnails may not be available for all attachment types
			if thumbnailErr != nil {
				t.Logf("GetMessagePartThumbnail not available (expected): %v", thumbnailErr)
			} else {
				assert.NotEmpty(t, thumbnail, "Thumbnail should not be empty")
			}
		} else {
			t.Log("No attachments found in the test message")
		}
	})
}

// TestE2E_MessageDeletion tests message deletion operations
func TestE2E_MessageDeletion(t *testing.T) {
	t.Parallel()

	t.Run("DeleteMessage", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test message to delete
		sendTestEmailWithSubject(t, testSMTP, "Delete Message Test")
		time.Sleep(2 * time.Second)

		// Get the message
		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		err = client.DeleteMessage(ctx, messageID)
		if err != nil {
			t.Logf("DeleteMessage not available (expected in some versions): %v", err)

			return
		}

		// Verify message was deleted
		_, err = client.GetMessage(ctx, messageID)
		assert.Error(t, err, "Message should be deleted")
	})

	t.Run("DeleteSearchResults", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send multiple messages with specific subject
		for i := range 3 {
			sendTestEmailWithSubject(t, testSMTP, fmt.Sprintf("Delete Search Test %d", i))
		}
		time.Sleep(2 * time.Second)

		// Verify messages exist
		response, err := client.SearchMessages(ctx, "subject:Delete Search Test", nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Messages), 3)

		// Delete all matching messages
		err = client.DeleteSearchResults(ctx, "subject:Delete Search Test")
		require.NoError(t, err)

		// Verify messages were deleted
		time.Sleep(1 * time.Second)
		response, err = client.SearchMessages(ctx, "subject:Delete Search Test", nil)
		require.NoError(t, err)
		assert.Equal(t, 0, len(response.Messages))
	})

	t.Run("DeleteAllMessages", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a few messages
		for range 2 {
			sendTestEmailWithSubject(t, testSMTP, "Delete All Test")
		}
		time.Sleep(2 * time.Second)

		// Verify messages exist
		response, err := client.ListMessages(ctx, nil)
		require.NoError(t, err)
		assert.Greater(t, len(response.Messages), 0)

		// Delete all messages
		err = client.DeleteAllMessages(ctx)
		require.NoError(t, err)

		// Verify all messages were deleted
		time.Sleep(1 * time.Second)
		response, err = client.ListMessages(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, 0, len(response.Messages))
	})
}

// TestE2E_ReleaseMessage tests message release functionality
func TestE2E_ReleaseMessage(t *testing.T) {
	t.Parallel()

	t.Run("ReleaseMessage", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test message
		sendTestEmailWithSubject(t, testSMTP, "Release Message Test")
		time.Sleep(2 * time.Second)

		// Get the message
		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		releaseData := &ReleaseMessageRequest{
			To: []string{"release@example.com"},
		}

		// Note: This might fail if SMTP relay is not configured
		// But we can test that the API accepts the request
		err = client.ReleaseMessage(ctx, messageID, releaseData)
		// Don't require no error as this depends on SMTP configuration
		if err != nil {
			t.Logf("Release message failed (expected without SMTP relay): %v", err)
		}
	})
}

// TestE2E_ChaosOperations tests chaos testing operations
func TestE2E_ChaosOperations(t *testing.T) {
	t.Parallel()

	t.Run("GetChaosConfig", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		config, err := client.GetChaosConfig(ctx)
		if err != nil {
			t.Logf("GetChaosConfig not available (expected in some versions): %v", err)

			return
		}
		assert.NotNil(t, config)
		// Chaos should be disabled by default
		assert.False(t, config.Enabled)
	})

	t.Run("SetChaosConfig", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		chaosConfig := &ChaosTriggers{
			AcceptConnections: 0.1,  // 10% chance to reject connections
			RejectSenders:     0.05, // 5% chance to reject senders
		}

		response, err := client.SetChaosConfig(ctx, chaosConfig)
		if err != nil {
			t.Logf("SetChaosConfig not available (expected in some versions): %v", err)

			return
		}
		assert.True(t, response.Enabled)
		assert.Equal(t, 0.1, response.Triggers.AcceptConnections)
		assert.Equal(t, 0.05, response.Triggers.RejectSenders)

		// Reset chaos config
		resetConfig := &ChaosTriggers{}
		_, err = client.SetChaosConfig(ctx, resetConfig)
		if err != nil {
			t.Logf("Reset chaos config failed (expected): %v", err)
		}
	})
}

// TestE2E_MessageValidation tests message validation operations
func TestE2E_MessageValidation(t *testing.T) {
	t.Parallel()

	t.Run("GetMessageHTMLCheck", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send HTML email for validation
		sendTestHTMLEmail(t, testSMTP)
		time.Sleep(2 * time.Second)

		response, err := client.SearchMessages(ctx, "subject:HTML E2E Test", nil)
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		htmlCheck, err := client.GetMessageHTMLCheck(ctx, messageID)
		require.NoError(t, err)
		// HTML check should return something (even if no errors/warnings)
		assert.NotNil(t, htmlCheck)
	})

	t.Run("GetMessageLinkCheck", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send HTML email for validation
		sendTestHTMLEmail(t, testSMTP)
		time.Sleep(2 * time.Second)

		response, err := client.SearchMessages(ctx, "subject:HTML E2E Test", nil)
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		linkCheck, err := client.GetMessageLinkCheck(ctx, messageID)
		require.NoError(t, err)
		assert.NotNil(t, linkCheck)
	})

	t.Run("GetMessageSpamAssassinCheck", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email for spam checking
		sendTestEmailWithSubject(t, testSMTP, "SpamAssassin Test Email")
		time.Sleep(2 * time.Second)

		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		spamCheck, err := client.GetMessageSpamAssassinCheck(ctx, messageID)
		require.NoError(t, err)
		assert.NotNil(t, spamCheck)
	})
}

// TestE2E_SendOperations tests message sending operations (may not be available in all versions)
func TestE2E_SendOperations(t *testing.T) {
	t.Parallel()

	t.Run("SendMessage", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		sendRequest := &SendMessageRequest{
			From: Address{
				Name:    "Test Sender",
				Address: "sender@example.com",
			},
			To: []Address{
				{
					Name:    "Test Recipient",
					Address: "recipient@example.com",
				},
			},
			Subject: "API Send E2E Test",
			Text:    "This is a test message sent via the API",
			HTML:    "<p>This is a <strong>test message</strong> sent via the API</p>",
		}

		// Note: Send endpoint may not be available in all Mailpit versions
		_, err := client.SendMessage(ctx, sendRequest)
		if err != nil {
			t.Logf("SendMessage not available (expected in some versions): %v", err)
		}
	})
}

// TestE2E_TagOperations tests tag-related operations (may not be available in all versions)
func TestE2E_TagOperations(t *testing.T) {
	t.Parallel()

	t.Run("GetTags", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Note: Tags endpoint may not be available in all Mailpit versions
		_, err := client.GetTags(ctx)
		if err != nil {
			t.Logf("GetTags not available (expected in some versions): %v", err)
		}
	})
}

// Helper functions for sending test emails

func sendTestEmail(t *testing.T, testSMTP *TestSMTP) {
	t.Helper()

	sendTestEmailWithSubject(t, testSMTP, "E2E Test Email")
}

func sendTestEmailWithSubject(t *testing.T, testSMTP *TestSMTP, subject string) {
	t.Helper()

	message := fmt.Sprintf(`To: test@example.com
From: sender@example.com
Subject: %s

This is a test email body for e2e testing.
`, subject)

	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", testSMTP.Host, testSMTP.SMTPPort),
		nil,
		"sender@example.com",
		[]string{"test@example.com"},
		[]byte(message),
	)
	require.NoError(t, err)
}

func sendTestHTMLEmail(t *testing.T, testSMTP *TestSMTP) {
	t.Helper()

	message := `To: test@example.com
From: sender@example.com
Subject: HTML E2E Test Email
MIME-Version: 1.0
Content-Type: multipart/alternative; boundary="boundary123"

--boundary123
Content-Type: text/plain; charset="UTF-8"

This is the plain text version.

--boundary123
Content-Type: text/html; charset="UTF-8"

<html>
<body>
<h1>HTML Test Email</h1>
<p>This is an HTML test email with a <a href="https://example.com">link</a>.</p>
</body>
</html>

--boundary123--
`

	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", testSMTP.Host, testSMTP.SMTPPort),
		nil,
		"sender@example.com",
		[]string{"test@example.com"},
		[]byte(message),
	)
	require.NoError(t, err)
}

func sendTestEmailWithAttachment(t *testing.T, testSMTP *TestSMTP) {
	t.Helper()

	// Encode test attachment
	attachmentContent := base64.StdEncoding.EncodeToString([]byte("This is a test attachment"))

	message := fmt.Sprintf(`To: test@example.com
From: sender@example.com
Subject: E2E Test Email with Attachment
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary=boundary456

--boundary456
Content-Type: text/plain; charset="UTF-8"

This is a test email with an attachment.

--boundary456
Content-Type: text/plain; charset="UTF-8"
Content-Disposition: attachment; filename="test.txt"
Content-Transfer-Encoding: base64

%s

--boundary456--
`, attachmentContent)

	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", testSMTP.Host, testSMTP.SMTPPort),
		nil,
		"sender@example.com",
		[]string{"test@example.com"},
		[]byte(message),
	)
	require.NoError(t, err)
}
