# Mailpit Go API Client

[![Go Reference](https://pkg.go.dev/badge/github.com/CodeLieutenant/mailpitclient.svg)](https://pkg.go.dev/github.com/CodeLieutenant/mailpitclient)
[![Go Report Card](https://goreportcard.com/badge/github.com/CodeLieutenant/mailpitclient)](https://goreportcard.com/report/github.com/CodeLieutenant/mailpitclient)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![codecov](https://codecov.io/gh/CodeLieutenant/mailpitclient/graph/badge.svg?token=lqhmgWPlWJ)](https://codecov.io/gh/CodeLieutenant/mailpitclient)


A **production-ready** Go client library for the [Mailpit](https://mailpit.axllent.org/) API, providing 100% coverage of all Mailpit API endpoints. Mailpit is a popular email testing tool with a REST API for managing emails, messages, tags, and server operations.

## ✨ Features

- 🚀 **Production-ready** with comprehensive error handling and retry logic
- 📡 **100% API coverage** - All Mailpit endpoints implemented and tested
- 🔄 **Context support** for cancellation and timeouts
- 🔒 **TLS/HTTPS support** with mkcert integration for testing
- ⚡ **High performance** with connection pooling and optimizations
- 🧪 **Comprehensive testing** with unit tests and E2E testing via testcontainers
- 🔧 **Thread-safe** - Safe for concurrent use across goroutines
- 📚 **Well-documented** with extensive examples and godoc comments

## 📋 Requirements

- Go 1.25.0 or later
- Mailpit server (for testing/production use)

## 🚀 Installation

```bash
go get github.com/CodeLieutenant/mailpitclient
```

## 🏃 Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    mailpit "github.com/CodeLieutenant/mailpitclient"
)

func main() {
    // Create client with default configuration (localhost:8025)
    client, err := mailpit.NewClient(nil)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Check server health
    if err := client.HealthCheck(ctx); err != nil {
        log.Fatal("Mailpit server not accessible:", err)
    }

    // Get server information
    info, err := client.GetServerInfo(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Connected to Mailpit v%s\n", info.Version)

    // List all messages
    messages, err := client.ListMessages(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d messages\n", messages.Total)
}
```

### Custom Configuration

```go
config := &mailpit.Config{
    BaseURL:    "https://mailpit.example.com",
    Timeout:    30 * time.Second,
    MaxRetries: 3,
    RetryDelay: 1 * time.Second,
    APIKey:     "your-api-key", // For authenticated instances
}

client, err := mailpit.NewClient(config)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

## 📖 Usage Examples

### Message Operations

#### List Messages with Pagination

```go
opts := &mailpit.ListOptions{
    Start: 0,
    Limit: 10,
}
messages, err := client.ListMessages(ctx, opts)
if err != nil {
    log.Fatal(err)
}

for _, msg := range messages.Messages {
    fmt.Printf("ID: %s, Subject: %s, From: %s\n",
        msg.ID, msg.Subject, msg.From.Address)
}
```

#### Get Message Details

```go
message, err := client.GetMessage(ctx, "message-id")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Subject: %s\n", message.Subject)
fmt.Printf("From: %s <%s>\n", message.From.Name, message.From.Address)
fmt.Printf("Date: %s\n", message.Date.Format(time.RFC3339))
fmt.Printf("HTML Body: %s\n", message.HTML)
fmt.Printf("Text Body: %s\n", message.Text)
```

#### Search Messages

```go
// Search by sender email
results, err := client.SearchMessages(ctx, "from:test@example.com", nil)
if err != nil {
    log.Fatal(err)
}

// Search with additional options
searchOpts := &mailpit.SearchOptions{
    Start: 0,
    Limit: 20,
}
results, err = client.SearchMessages(ctx, "subject:important", searchOpts)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d matching messages\n", results.Total)
```

#### Message Analysis

```go
// HTML validation
htmlCheck, err := client.GetMessageHTMLCheck(ctx, "message-id")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("HTML errors: %d, warnings: %d\n",
    len(htmlCheck.Errors), len(htmlCheck.Warnings))

// Link validation
linkCheck, err := client.GetMessageLinkCheck(ctx, "message-id")
if err != nil {
    log.Fatal(err)
}
for _, link := range linkCheck.Links {
    fmt.Printf("URL: %s, Status: %d, Valid: %v\n",
        link.URL, link.Status, link.Valid)
}

// SpamAssassin analysis
saCheck, err := client.GetMessageSpamAssassinCheck(ctx, "message-id")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Spam score: %.2f, Required: %.2f\n",
    saCheck.Score, saCheck.RequiredScore)
```

### Send Operations

```go
message := &mailpit.SendMessageRequest{
    From: mailpit.Address{
        Address: "sender@example.com",
        Name:    "Test Sender",
    },
    To: []mailpit.Address{
        {Address: "recipient@example.com", Name: "Test Recipient"},
    },
    Cc: []mailpit.Address{
        {Address: "cc@example.com", Name: "CC Recipient"},
    },
    Subject: "Test Message from Go Client",
    HTML:    "<h1>Hello World</h1><p>This is a <strong>test</strong> message.</p>",
    Text:    "Hello World\n\nThis is a test message.",
    Tags:    []string{"test", "automated", "go-client"},
}

result, err := client.SendMessage(ctx, message)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Message sent with ID: %s\n", result.ID)
```

### Tag Operations

```go
// Get all available tags
tags, err := client.GetTags(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Available tags: %v\n", tags)

// Set global tags
newTags := []string{"important", "test", "automated", "production"}
updatedTags, err := client.SetTags(ctx, newTags)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Updated global tags: %v\n", updatedTags)

// Tag specific messages
messageIDs := []string{"msg-1", "msg-2", "msg-3"}
err = client.SetMessageTags(ctx, "urgent", messageIDs)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Messages tagged as 'urgent'")

// Delete a tag
err = client.DeleteTag(ctx, "old-tag")
if err != nil {
    log.Fatal(err)
}
```

### Server Operations

```go
// Get comprehensive server information
info, err := client.GetServerInfo(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Mailpit Version: %s\n", info.Version)
fmt.Printf("Database: %s\n", info.Database)

// Get server statistics
stats, err := client.GetStats(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Total messages: %d\n", stats.Total)
fmt.Printf("Unread messages: %d\n", stats.Unread)

// Get Web UI configuration
webConfig, err := client.GetWebUIConfig(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Read-only mode: %v\n", webConfig.ReadOnly)
fmt.Printf("SMTP server enabled: %v\n", webConfig.SMTPEnabled)
```

### Advanced Features

#### Message Release (SMTP Relay)

```go
releaseReq := &mailpit.ReleaseMessageRequest{
    To:   []string{"recipient@production.com"},
    Host: "smtp.gmail.com",
    Port: 587,
    Auth: &mailpit.SMTPAuth{
        Username: "your-email@gmail.com",
        Password: "app-password",
        AuthType: "PLAIN",
    },
    TLS: true,
}

err := client.ReleaseMessage(ctx, "message-id", releaseReq)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Message released via SMTP")
```

#### Chaos Testing (for resilience testing)

```go
// Get current chaos configuration
chaosConfig, err := client.GetChaosConfig(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Chaos testing enabled: %v\n", chaosConfig.Enabled)

// Configure chaos triggers
triggers := &mailpit.ChaosTriggers{
    AcceptConnections: 0.9,  // 90% success rate
    RejectSenders:     0.1,  // 10% sender rejection
    DelayConnections:  0.2,  // 20% connection delay
    DelayDuration:     5,    // 5 second delays
}

updated, err := client.SetChaosConfig(ctx, triggers)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Chaos triggers updated: %+v\n", updated)
```

### Error Handling

```go
message, err := client.GetMessage(ctx, "invalid-id")
if err != nil {
    var mailpitErr *mailpit.Error
    if errors.As(err, &mailpitErr) {
        switch mailpitErr.Type {
        case mailpit.ErrorTypeAPI:
            if mailpitErr.StatusCode == 404 {
                fmt.Println("Message not found")
            } else {
                fmt.Printf("API error %d: %s\n", mailpitErr.StatusCode, mailpitErr.Message)
            }
        case mailpit.ErrorTypeNetwork:
            fmt.Println("Network error:", mailpitErr.Message)
        case mailpit.ErrorTypeValidation:
            fmt.Println("Validation error:", mailpitErr.Message)
        case mailpit.ErrorTypeTimeout:
            fmt.Println("Request timed out:", mailpitErr.Message)
        }
    } else {
        fmt.Println("Unknown error:", err)
    }
}
```

## 🧪 Testing Integration

This client is designed for seamless integration with [testcontainers](https://testcontainers.com/) for comprehensive testing:

```go
func TestWithMailpit(t *testing.T) {
    ctx := context.Background()

    // Start Mailpit container
    container, err := testcontainers.GenericContainer(ctx,
        testcontainers.GenericContainerRequest{
            ContainerRequest: testcontainers.ContainerRequest{
                Image:        "axllent/mailpit:latest",
                ExposedPorts: []string{"8025/tcp", "1025/tcp"},
                WaitingFor:   wait.ForHTTP("/api/v1/info").WithPort("8025/tcp"),
                Env: map[string]string{
                    "MP_SMTP_AUTH_ACCEPT_ANY": "1",
                    "MP_SMTP_AUTH_ALLOW_INSECURE": "1",
                },
            },
            Started: true,
        })
    require.NoError(t, err)
    defer container.Terminate(ctx)

    // Get container connection details
    host, err := container.Host(ctx)
    require.NoError(t, err)

    httpPort, err := container.MappedPort(ctx, "8025")
    require.NoError(t, err)

    smtpPort, err := container.MappedPort(ctx, "1025")
    require.NoError(t, err)

    // Create API client
    client, err := mailpit.NewClient(&mailpit.Config{
        BaseURL: fmt.Sprintf("http://%s:%s", host, httpPort.Port()),
        Timeout: 10 * time.Second,
    })
    require.NoError(t, err)
    defer client.Close()

    // Send test email via SMTP
    smtpClient, err := smtp.Dial(fmt.Sprintf("%s:%s", host, smtpPort.Port()))
    require.NoError(t, err)
    defer smtpClient.Close()

    err = smtpClient.Mail("test@example.com")
    require.NoError(t, err)

    err = smtpClient.Rcpt("recipient@example.com")
    require.NoError(t, err)

    writer, err := smtpClient.Data()
    require.NoError(t, err)

    _, err = writer.Write([]byte(`Subject: Test Email
From: test@example.com
To: recipient@example.com

This is a test email body.`))
    require.NoError(t, err)
    require.NoError(t, writer.Close())

    // Wait for message processing
    time.Sleep(2 * time.Second)

    // Verify message received
    messages, err := client.ListMessages(ctx, nil)
    require.NoError(t, err)
    require.Equal(t, 1, messages.Total)

    message := messages.Messages[0]
    assert.Equal(t, "Test Email", message.Subject)
    assert.Equal(t, "test@example.com", message.From.Address)
}
```

## 📊 API Coverage

This client provides **100% coverage** of the Mailpit API endpoints. For detailed endpoint mapping and implementation status, see our [API Coverage Documentation](API_COVERAGE.md).

### Automated Coverage Testing

We maintain automated testing to ensure complete API coverage:

```bash
# Run full API coverage validation
make test-api-coverage

# Run fast offline coverage test
make test-api-coverage-offline
```

## 🛠️ Development

### Prerequisites

- Go 1.25.0+
- Docker (for testing)
- Make
- [golangci-lint](https://golangci-lint.run/) (for linting)

### Development Commands

```bash
# Install dependencies
go mod download

# Run tests
make test

# Run linting
make check

# Auto-fix lint issues
make fix

# Run field alignment optimization
make fieldalign

# Security scanning
make security

# Generate TLS certificates for testing
make mkcert-generate HOSTS='localhost 127.0.0.1 ::1'
```

### Testing

The project includes comprehensive testing:

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: End-to-end testing with real Mailpit instances
- **API Coverage Tests**: Automated verification of complete API coverage

```bash
# Run all tests
make test

# Run with coverage
go test -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out
```

## 📚 Documentation

- [API Coverage Documentation](API_COVERAGE.md) - Complete endpoint mapping
- [GoDoc](https://pkg.go.dev/github.com/CodeLieutenant/mailpitclient) - Comprehensive API documentation
- [Mailpit Documentation](https://mailpit.axllent.org/) - Official Mailpit documentation

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Run** tests and linting (`make check test`)
5. **Push** to the branch (`git push origin feature/amazing-feature`)
6. **Open** a Pull Request

### Code Standards

- Follow the established patterns in the codebase
- Add tests for new functionality (unit + E2E)
- All tests must use `t.Parallel()` for parallel execution
- Maintain >90% code coverage
- Use `make check` to validate code quality
- Update documentation as needed

## 📄 License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.

## 🔗 Related Projects

- [Mailpit](https://github.com/axllent/mailpit) - The email testing tool this client connects to
- [Testcontainers Go](https://github.com/testcontainers/testcontainers-go) - Docker testing framework used in our test suite

## ⭐ Support

If you find this project helpful, please consider giving it a star on GitHub!

For bugs, feature requests, or questions, please [open an issue](https://github.com/CodeLieutenant/mailpitclient/issues).

---

*Built with ❤️ for the Go community*
