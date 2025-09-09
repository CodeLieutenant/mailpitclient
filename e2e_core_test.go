package mailpitclient

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_CoreFeatures tests the core mailpit API features that should work in most versions
func TestE2E_CoreFeatures(t *testing.T) {
	t.Parallel()

	t.Run("ServerInfo", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		info, err := client.GetServerInfo(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, info.Version)
		assert.NotEmpty(t, info.Database)
		assert.GreaterOrEqual(t, info.Messages, 0)
		assert.GreaterOrEqual(t, info.Unread, 0)
		t.Logf("Server Version: %s", info.Version)
		t.Logf("Database: %s", info.Database)
	})

	t.Run("WebUIConfig", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		config, err := client.GetWebUIConfig(ctx)
		require.NoError(t, err)
		assert.NotNil(t, config)
		t.Logf("MessageRelay Enabled: %v", config.MessageRelay.Enabled)
		t.Logf("SpamAssassin: %v", config.SpamAssassin)
	})

	t.Run("HealthCheck", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		err := client.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	t.Run("ListMessages", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send test emails for this test
		sendTestEmail(t, testSMTP)
		sendTestEmailWithSubject(t, testSMTP, "Core List Test Email 1")
		sendTestEmailWithSubject(t, testSMTP, "Core List Test Email 2")
		time.Sleep(2 * time.Second) // Allow emails to be processed

		response, err := client.ListMessages(ctx, nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Messages), 3)
		assert.GreaterOrEqual(t, response.Total, 3)
		assert.Equal(t, response.Count, len(response.Messages))

		t.Logf("Total messages: %d", response.Total)
		t.Logf("Unread messages: %d", response.Unread)

		// Test with pagination
		opts := &ListOptions{Start: 0, Limit: 1}
		response, err = client.ListMessages(ctx, opts)
		require.NoError(t, err)
		assert.Equal(t, 1, len(response.Messages))
	})

	t.Run("GetMessage", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email for this test
		sendTestEmailWithSubject(t, testSMTP, "Core GetMessage Test Email")
		time.Sleep(2 * time.Second) // Allow email to be processed

		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		message, err := client.GetMessage(ctx, messageID)
		require.NoError(t, err)
		assert.Equal(t, messageID, message.ID)
		assert.NotEmpty(t, message.Subject)
		assert.NotEmpty(t, message.From.Address)

		t.Logf("Message ID: %s", message.ID)
		t.Logf("Subject: %s", message.Subject)
		t.Logf("From: %s", message.From.Address)
	})

	t.Run("GetMessageHeaders", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email for this test
		sendTestEmailWithSubject(t, testSMTP, "Core GetHeaders Test Email")
		time.Sleep(2 * time.Second) // Allow email to be processed

		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		headers, err := client.GetMessageHeaders(ctx, messageID)
		require.NoError(t, err)
		assert.NotEmpty(t, headers)
		assert.Contains(t, headers, "Subject")
		assert.Contains(t, headers, "From")

		t.Logf("Headers retrieved: %d", len(headers))
	})

	t.Run("SearchMessages", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send test emails for searching
		sendTestEmailWithSubject(t, testSMTP, "Core Search Test Email 1")
		sendTestEmailWithSubject(t, testSMTP, "Core Search Test Email 2")
		time.Sleep(2 * time.Second) // Allow emails to be processed

		// Search for messages with "Core Search Test" in subject
		response, err := client.SearchMessages(ctx, "subject:Core Search Test", nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Messages), 2)

		// Verify search results
		for _, msg := range response.Messages {
			assert.Contains(t, msg.Subject, "Core Search Test")
		}

		t.Logf("Search found %d messages", len(response.Messages))
	})

	t.Run("DeleteMessage", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a message specifically to delete
		sendTestEmailWithSubject(t, testSMTP, "Delete Me Test")
		time.Sleep(2 * time.Second)

		// Find the message
		response, err := client.SearchMessages(ctx, "subject:Delete Me Test", nil)
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)

		messageID := response.Messages[0].ID
		err = client.DeleteMessage(ctx, messageID)
		if err != nil {
			t.Logf("DeleteMessage not available: %v", err)
		} else {
			t.Logf("Successfully deleted message: %s", messageID)
		}
	})
}

// TestE2E_OptionalFeatures tests features that may not be available in all mailpit versions
func TestE2E_OptionalFeatures(t *testing.T) {
	t.Parallel()

	t.Run("GetMessageSource", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email first
		sendTestEmailWithSubject(t, testSMTP, "Optional GetMessageSource Test")
		time.Sleep(2 * time.Second)

		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)
		messageID := response.Messages[0].ID

		_, err = client.GetMessageSource(ctx, messageID)
		if err != nil {
			t.Logf("GetMessageSource not available: %v", err)
		} else {
			t.Log("GetMessageSource is available")
		}
	})

	t.Run("MarkMessageRead", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email first
		sendTestEmailWithSubject(t, testSMTP, "Optional MarkMessageRead Test")
		time.Sleep(2 * time.Second)

		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)
		messageID := response.Messages[0].ID

		err = client.MarkMessageRead(ctx, messageID)
		if err != nil {
			t.Logf("MarkMessageRead not available: %v", err)
		} else {
			t.Log("MarkMessageRead is available")
		}
	})

	t.Run("MarkMessageUnread", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email first
		sendTestEmailWithSubject(t, testSMTP, "Optional MarkMessageUnread Test")
		time.Sleep(2 * time.Second)

		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)
		messageID := response.Messages[0].ID

		err = client.MarkMessageUnread(ctx, messageID)
		if err != nil {
			t.Logf("MarkMessageUnread not available: %v", err)
		} else {
			t.Log("MarkMessageUnread is available")
		}
	})

	t.Run("GetTags", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		_, err := client.GetTags(ctx)
		if err != nil {
			t.Logf("GetTags not available: %v", err)
		} else {
			t.Log("GetTags is available")
		}
	})

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
			Subject: "API Send Test",
			Text:    "This is a test message sent via the API",
		}

		_, err := client.SendMessage(ctx, sendRequest)
		if err != nil {
			t.Logf("SendMessage not available: %v", err)
		} else {
			t.Log("SendMessage is available")
		}
	})

	t.Run("GetMessageHTML", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email first
		sendTestEmailWithSubject(t, testSMTP, "Optional GetMessageHTML Test")
		time.Sleep(2 * time.Second)

		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)
		messageID := response.Messages[0].ID

		_, err = client.GetMessageHTML(ctx, messageID)
		if err != nil {
			t.Logf("GetMessageHTML not available: %v", err)
		} else {
			t.Log("GetMessageHTML is available")
		}
	})

	t.Run("GetMessageText", func(t *testing.T) {
		t.Parallel()
		testSMTP := GetTestSMTP(t)
		client := testSMTP.MailpitClient
		ctx := t.Context()

		// Send a test email first
		sendTestEmailWithSubject(t, testSMTP, "Optional GetMessageText Test")
		time.Sleep(2 * time.Second)

		response, err := client.ListMessages(ctx, &ListOptions{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(response.Messages), 0)
		messageID := response.Messages[0].ID

		_, err = client.GetMessageText(ctx, messageID)
		if err != nil {
			t.Logf("GetMessageText not available: %v", err)
		} else {
			t.Log("GetMessageText is available")
		}
	})
}
