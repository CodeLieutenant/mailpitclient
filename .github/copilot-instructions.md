# Copilot Instructions for mailpit-go-api

This is a production-ready Go client library for the Mailpit API, designed for high performance and reliability in production environments. Follow these guidelines when generating code for this project.

## Project Overview

The `mailpit-go-api` is a comprehensive Go client library that provides 100% coverage of the Mailpit API endpoints. Mailpit is an email testing tool with a REST API for managing emails, messages, tags, and server operations.

**Key Features:**
- Complete API coverage with all Mailpit endpoints
- Production-ready client with proper error handling
- Comprehensive unit and E2E testing
- Performance-optimized with context support
- Secure TLS/HTTPS support with mkcert integration
- Docker-based testing with testcontainers

## Code Quality Standards

### Linting and Code Quality
- **ALWAYS** use `go tool golangci-lint` for linting
- Run `make check` before committing any code
- Use `make fix` to auto-fix lint issues where possible
- Apply `make fieldalign` for struct field alignment optimization
- Run `make security` (gosec) for security vulnerability checks

### Go Best Practices
1. **Error Handling**: Use explicit error handling with detailed error messages
2. **Context**: Always use `context.Context` for API calls and cancellation
3. **Performance**:
   - Use struct field alignment for memory efficiency
   - Implement connection pooling and reuse
   - Use buffered I/O for large data transfers
   - Optimize JSON marshaling/unmarshaling
4. **Concurrency**: Use goroutines safely with proper synchronization
5. **Memory Management**: Avoid memory leaks, close resources properly
6. **Type Safety**: Use strong typing, avoid `interface{}` when possible

### Code Structure
- Follow the established pattern in existing files
- Maintain separation of concerns: client, types, errors, tests
- Use interfaces for testability and extensibility
- Implement proper resource cleanup with `defer` statements

## Testing Requirements

### Testing Guidelines - MANDATORY PARALLEL EXECUTION

- **EVERY test function MUST have `t.Parallel()` as the first line after the test declaration**
- **EVERY subtest MUST have `t.Parallel()` as the first line after the subtest declaration**
- **Each test and subtest MUST have its own isolated test environment** - call `GetTestSMTP(t)` within each test/subtest function, never share across tests
- **NO shared state between tests** - tests should be completely independent and able to run in any order
- **Use proper test isolation** - send specific test emails within each test rather than relying on emails sent by other tests
- **Include proper wait times** - use `time.Sleep(2 * time.Second)` after sending emails to allow processing
- **Use unique test subjects** - use different subjects for each test to avoid conflicts (e.g., "GetMessage Test Email", "SearchMessages Test Email")

### Example Test Pattern - FOLLOW THIS EXACTLY

```go
func TestExample(t *testing.T) {
    t.Parallel()  // REQUIRED - FIRST LINE

    t.Run("SubTest", func(t *testing.T) {
        t.Parallel()  // REQUIRED - FIRST LINE
        testSMTP := GetTestSMTP(t)  // REQUIRED - get own instance
        client := testSMTP.MailpitClient
        ctx := t.Context()

        // Send test data specific to this test
        sendTestEmailWithSubject(t, testSMTP, "SubTest Email")
        time.Sleep(2 * time.Second)

        // Test logic...
    })
}
```

### Race Condition Prevention - CRITICAL

- Never share variables between parallel tests
- Always use local variables within test functions
- Each test gets its own SMTP container and client
- Use proper synchronization when needed
- Avoid global state or shared resources
- All tests MUST pass with `-race` flag

### Unit Tests
- **EVERY** public function MUST have unit tests
- Use table-driven tests for multiple test cases
- Test both success and failure scenarios
- Mock external dependencies
- Achieve >90% code coverage
- Use `testing.T` and `testify/require` for assertions

Example unit test structure:
```go
func TestClient_MethodName(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name           string
        input          InputType
        mockSetup      func(*MockClient)
        expectedResult ExpectedType
        expectedError  string
    }{
        // test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            // test implementation
        })
    }
}
```

### End-to-End Tests
- **EVERY** API endpoint MUST have E2E tests
- Use testcontainers for real Mailpit server testing
- Test complete workflows and integrations
- Use `GetTestSMTP()` helper for test setup
- Clean up test data with `ClearMessages()`
- Wait for async operations with `WaitForMessages()`

### Test Organization
- Unit tests in `*_test.go` files alongside source files
- E2E tests in `e2e_*_test.go` files
- Core E2E tests in `e2e_core_test.go`
- Test helpers in `testhelpers_test.go`

## API Client Implementation

### Client Interface
The main `Client` interface defines all available operations:
- Message operations (CRUD, search, attachments)
- Server operations (health, info, config)
- Tag operations (get, set, delete)
- View operations (HTML, text, raw content)
- Send operations (compose and send emails)

### Error Handling
- Use custom error types defined in `errors.go`
- Wrap errors with context using `fmt.Errorf("operation failed: %w", err)`
- Handle HTTP status codes appropriately
- Provide meaningful error messages for debugging

### HTTP Client Configuration
- Use custom HTTP client with proper timeouts
- Implement retry logic for transient failures
- Support connection pooling and keep-alive
- Handle TLS configuration for HTTPS endpoints
- Set appropriate User-Agent headers

## Mailpit API Endpoints Coverage

The client must implement ALL these endpoints with corresponding methods:

### Core Message Operations
- `GET /api/v1/messages` → `ListMessages()`
- `DELETE /api/v1/messages` → `DeleteAllMessages()`
- `GET /api/v1/message/{id}` → `GetMessage()`
- `DELETE /api/v1/message/{id}` → `DeleteMessage()`
- `GET /api/v1/message/{id}/headers` → `GetMessageHeaders()`
- `GET /api/v1/message/{id}/source` → `GetMessageSource()`
- `GET /api/v1/message/{id}/events` → `GetMessageEvents()`
- `POST /api/v1/message/{id}/release` → `ReleaseMessage()`
- `PUT /api/v1/messages/{id}/read` → `MarkMessageRead()`
- `PUT /api/v1/messages/{id}/unread` → `MarkMessageUnread()`

### Message Content Operations
- `GET /api/v1/view/{id}/html` → `GetMessageHTML()`
- `GET /api/v1/view/{id}/text` → `GetMessageText()`
- `GET /api/v1/view/{id}/raw` → `GetMessageRaw()`
- `GET /api/v1/view/{id}/part/{partId}/html` → `GetMessagePartHTML()`
- `GET /api/v1/view/{id}/part/{partId}/text` → `GetMessagePartText()`

### Message Parts & Attachments
- `GET /api/v1/message/{id}/part/{partId}` → `GetMessagePart()`
- `GET /api/v1/message/{id}/part/{partId}/thumb` → `GetMessagePartThumbnail()`
- `GET /api/v1/message/{id}/attachment/{attachmentId}` → `GetMessageAttachment()`

### Message Analysis
- `GET /api/v1/message/{id}/html-check` → `GetMessageHTMLCheck()`
- `GET /api/v1/message/{id}/link-check` → `GetMessageLinkCheck()`
- `GET /api/v1/message/{id}/sa-check` → `GetMessageSpamAssassinCheck()`

### Search Operations
- `GET /api/v1/search` → `SearchMessages()`
- `DELETE /api/v1/search` → `DeleteSearchResults()`

### Send Operations
- `POST /api/v1/send` → `SendMessage()`

### Tag Operations
- `GET /api/v1/tags` → `GetTags()`
- `PUT /api/v1/tags` → `SetTags()`
- `PUT /api/v1/tags/{tag}` → `SetMessageTags()`
- `DELETE /api/v1/tags/{tag}` → `DeleteTag()`

### Server Operations
- `GET /api/v1/info` → `GetServerInfo()`
- `GET /api/v1/webui` → `GetWebUIConfig()`
- `GET /livez` → `HealthCheck()`

## Makefile Commands Reference

Use these commands for development and CI/CD:

### Code Quality & Linting
- `make check` - Run golangci-lint with full linting rules
- `make fix` - Auto-fix lint issues where possible
- `make fieldalign` - Apply field alignment optimizations
- `make fmt` - Format code using golangci-lint formatter
- `make tidy` - Tidy go.mod and download dependencies
- `make security` - Run gosec security vulnerability scanner

### Testing
- `make test` - Run unit tests with coverage (JSON output via gotestfmt)
  - Includes atomic coverage mode
  - Optimized for debugging with `-gcflags='all=-N -l'`
  - 5-minute timeout for comprehensive testing
  - Coverage output to `coverage.txt`

### TLS Certificate Management (mkcert)
- `make install-mkcert` - Install mkcert binary for local TLS certificates
- `make mkcert-install-ca` - Install mkcert CA into system trust stores
- `make mkcert-generate` - Generate TLS certificates for testing
  - Usage: `make mkcert-generate HOSTS='localhost 127.0.0.1 ::1'`
- `make mkcert-uninstall-ca` - Remove mkcert CA from system trust stores

## Performance Considerations

### Memory Optimization
- Use pointer receivers for large structs
- Implement proper struct field alignment
- Reuse buffers for repeated operations
- Close HTTP response bodies immediately after use

### Connection Management
- Implement connection pooling
- Use HTTP/2 when available
- Set appropriate timeouts and keep-alive settings
- Handle connection errors gracefully with retries

### Concurrency
- Use worker pools for bulk operations
- Implement proper context cancellation
- Use channels for goroutine communication
- Avoid data races with proper synchronization

## Documentation Standards

- Every public function must have godoc comments
- Include usage examples in doc comments
- Document error conditions and return values
- Maintain API coverage documentation in `API_COVERAGE.md`
- Update `doc.go` with comprehensive usage examples

## Security Considerations

- Validate all input parameters
- Use HTTPS by default in production
- Implement proper authentication if required
- Sanitize error messages to avoid information disclosure
- Use secure defaults for HTTP client configuration

## Type Definitions

Maintain strong typing throughout the codebase:
- Use specific types for API responses (`MessagesResponse`, `Message`, etc.)
- Define proper request/response structures
- Use enums for constants and status values
- Implement JSON marshaling/unmarshaling correctly

## Dependencies

The project uses these key dependencies:
- `github.com/stretchr/testify` - Testing framework and assertions
- `github.com/testcontainers/testcontainers-go` - Docker-based testing
- Standard library packages for HTTP, JSON, context, etc.

## Integration Testing

- Use Docker containers via testcontainers for real Mailpit instances
- Test against actual SMTP server functionality
- Verify email sending and receiving workflows
- Test TLS/SSL connections with generated certificates
- Validate API responses match expected schemas

When implementing new features or fixing bugs, ensure you follow all these guidelines and maintain the high-quality standards established in this production-ready client library.
