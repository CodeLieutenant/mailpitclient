// package mailpit_go_api provides a production-ready Go client for interacting with Mailpit API.
//
// Mailpit is a popular email testing tool that provides a REST API for managing emails.
// This client provides comprehensive functionality for managing messages, retrieving server
// information, and performing health checks.
//
// # Basic Usage
//
// Create a client with default configuration:
//
//	client, err := mailpit.NewClient(nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
// # Custom Configuration
//
// Create a client with custom configuration:
//
//	config := &mailpit.Config{
//		BaseURL:    "http://localhost:8025",
//		Timeout:    30 * time.Second,
//		MaxRetries: 3,
//		RetryDelay: 1 * time.Second,
//	}
//
//	client, err := mailpit.NewClient(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
// # Message Operations
//
// List all messages:
//
//	messages, err := client.ListMessages(ctx, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Found %d messages\n", messages.Total)
//
// List messages with pagination:
//
//	opts := &mailpit.ListOptions{
//		Start: 0,
//		Limit: 10,
//	}
//	messages, err := client.ListMessages(ctx, opts)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Get a specific message:
//
//	message, err := client.GetMessage(ctx, "message-id")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Subject: %s\n", message.Subject)
//
// Search messages:
//
//	results, err := client.SearchMessages(ctx, "test@example.com", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Found %d matching messages\n", len(results.Messages))
//
// Delete all messages:
//
//	err := client.DeleteAllMessages(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Get message headers:
//
//	headers, err := client.GetMessageHeaders(ctx, "message-id")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Content-Type: %v\n", headers["Content-Type"])
//
// Perform HTML validation on a message:
//
//	htmlCheck, err := client.GetMessageHTMLCheck(ctx, "message-id")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("HTML errors: %d, warnings: %d\n", len(htmlCheck.Errors), len(htmlCheck.Warnings))
//
// Check links in a message:
//
//	linkCheck, err := client.GetMessageLinkCheck(ctx, "message-id")
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, link := range linkCheck.Links {
//		fmt.Printf("URL: %s, Status: %d\n", link.URL, link.Status)
//	}
//
// Get SpamAssassin score for a message:
//
//	saCheck, err := client.GetMessageSpamAssassinCheck(ctx, "message-id")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Spam score: %.2f\n", saCheck.Score)
//
// Release a message via SMTP relay:
//
//	releaseReq := &mailpit.ReleaseMessageRequest{
//		To:   []string{"recipient@example.com"},
//		Host: "smtp.example.com",
//		Port: 587,
//	}
//	err := client.ReleaseMessage(ctx, "message-id", releaseReq)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # Server Operations
//
// Check server health:
//
//	err := client.HealthCheck(ctx)
//	if err != nil {
//		log.Fatal("Server is not healthy:", err)
//	}
//
// Get server information:
//
//	info, err := client.GetServerInfo(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Mailpit version: %s\n", info.Version)
//
// Get server statistics:
//
//	stats, err := client.GetStats(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Total messages: %d, Unread: %d\n", stats.Total, stats.Unread)
//
// Get web UI configuration:
//
//	config, err := client.GetWebUIConfig(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Read-only mode: %v\n", config.ReadOnly)
//
// # Send Operations
//
// Send a message via HTTP API:
//
//	message := &mailpit.SendMessageRequest{
//		From:    mailpit.Address{Address: "sender@example.com", Name: "Sender"},
//		To:      []mailpit.Address{{Address: "recipient@example.com", Name: "Recipient"}},
//		Subject: "Test Message",
//		HTML:    "<h1>Hello World</h1><p>This is a test message.</p>",
//		Text:    "Hello World\n\nThis is a test message.",
//		Tags:    []string{"test", "automated"},
//	}
//
//	result, err := client.SendMessage(ctx, message)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Message sent with ID: %s\n", result.ID)
//
// # Tags Operations
//
// Get all available tags:
//
//	tags, err := client.GetTags(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Available tags: %v\n", tags)
//
// Set tags for the server:
//
//	newTags := []string{"important", "test", "automated"}
//	updatedTags, err := client.SetTags(ctx, newTags)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Updated tags: %v\n", updatedTags)
//
// Tag specific messages:
//
//	messageIDs := []string{"msg-1", "msg-2", "msg-3"}
//	err := client.SetMessageTags(ctx, "important", messageIDs)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Delete a tag:
//
//	err := client.DeleteTag(ctx, "old-tag")
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # View Operations
//
// Get HTML view of a message:
//
//	htmlContent, err := client.GetMessageHTML(ctx, "message-id")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("HTML content length: %d\n", len(htmlContent))
//
// Get text view of a message:
//
//	textContent, err := client.GetMessageText(ctx, "message-id")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Text content: %s\n", textContent)
//
// # Chaos Testing Operations
//
// Get current chaos configuration:
//
//	chaosConfig, err := client.GetChaosConfig(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Chaos enabled: %v\n", chaosConfig.Enabled)
//
// Configure chaos triggers for testing:
//
//	triggers := &mailpit.ChaosTriggers{
//		AcceptConnections: 0.9,  // 90% success rate
//		RejectSenders:     0.1,  // 10% sender rejection
//		DelayConnections:  0.2,  // 20% connection delay
//	}
//
//	updated, err := client.SetChaosConfig(ctx, triggers)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Chaos triggers updated, enabled: %v\n", updated.Enabled)
//
// # Error Handling
//
// The client provides structured error handling with different error types:
//
//	_, err := client.GetMessage(ctx, "invalid-id")
//	if err != nil {
//		var mailpitErr *mailpit.Error
//		if errors.As(err, &mailpitErr) {
//			switch mailpitErr.Type {
//			case mailpit.ErrorTypeAPI:
//				if mailpitErr.StatusCode == 404 {
//					fmt.Println("Message not found")
//				}
//			case mailpit.ErrorTypeNetwork:
//				fmt.Println("Network error:", mailpitErr.Message)
//			case mailpit.ErrorTypeValidation:
//				fmt.Println("Validation error:", mailpitErr.Message)
//			}
//		}
//	}
//
// # Authentication
//
// If your Mailpit server requires authentication, configure it:
//
//	config := &mailpit.Config{
//		BaseURL:  "http://localhost:8025",
//		APIKey:   "your-api-key", // Bearer token authentication
//		// OR
//		Username: "user",         // Basic authentication
//		Password: "pass",
//	}
//
// # Testing Integration
//
// This client is designed to work seamlessly with testcontainers for integration testing:
//
//	// Start mailpit container
//	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
//		ContainerRequest: testcontainers.ContainerRequest{
//			Image:        "axllent/mailpit:latest",
//			ExposedPorts: []string{"8025/tcp"},
//			WaitingFor:   wait.ForHTTP("/api/v1/info").WithPort("8025/tcp"),
//		},
//		Started: true,
//	})
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer container.Terminate(ctx)
//
//	// Get container details and create client
//	host, _ := container.Host(ctx)
//	port, _ := container.MappedPort(ctx, "8025")
//
//	client, err := mailpit.NewClient(&mailpit.Config{
//		BaseURL: fmt.Sprintf("http://%s:%s", host, port.Port()),
//	})
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer client.Close()
//
//	// Use client in tests
//	err = client.DeleteAllMessages(ctx)
//	require.NoError(t, err)
//
// # Production Considerations
//
// For production use, consider:
//
// - Set appropriate timeouts based on your network conditions
// - Configure retry logic for transient failures
// - Use context with proper cancellation
// - Monitor client performance and adjust configuration
// - Implement proper logging and observability
//
// The client is thread-safe and can be used concurrently from multiple goroutines.
package mailpit_go_api
