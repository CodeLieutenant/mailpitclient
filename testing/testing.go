// Package testing provides testing utilities and infrastructure for mailpitclient end-to-end tests.
//
// This package offers a complete testing framework for the mailpitclient library, including
// Docker-based Mailpit server setup, SMTP container pooling, and helper functions for
// common testing scenarios.
//
// # Basic Usage
//
// The primary entry point is GetTestSMTP(), which provides a complete test environment:
//
//	func TestExample(t *testing.T) {
//		t.Parallel()  // REQUIRED for parallel execution
//
//		testSMTP := GetTestSMTP(t)
//		client := testSMTP.MailpitClient
//		ctx := t.Context()
//
//		// Your test code here
//		messages, err := client.ListMessages(ctx, nil)
//		require.NoError(t, err)
//	}
//
// # Testing Requirements
//
// MANDATORY: All tests MUST follow these parallel execution requirements:
//
// 1. Every test function MUST have t.Parallel() as the first line
// 2. Every subtest MUST have t.Parallel() as the first line
// 3. Each test/subtest MUST get its own TestSMTP instance via GetTestSMTP(t)
// 4. Never share TestSMTP instances between tests - this causes race conditions
//
// Example of proper parallel test structure:
//
//	func TestMailpitClient_Operations(t *testing.T) {
//		t.Parallel()  // REQUIRED - FIRST LINE
//
//		t.Run("ListMessages", func(t *testing.T) {
//			t.Parallel()  // REQUIRED - FIRST LINE
//			testSMTP := GetTestSMTP(t)  // REQUIRED - own instance
//			client := testSMTP.MailpitClient
//			ctx := t.Context()
//
//			// Send test email specific to this test
//			sendTestEmailWithSubject(t, testSMTP, "ListMessages Test Email")
//			time.Sleep(2 * time.Second)  // Allow processing time
//
//			messages, err := client.ListMessages(ctx, nil)
//			require.NoError(t, err)
//			require.GreaterOrEqual(t, len(messages.Messages), 1)
//		})
//
//		t.Run("GetMessage", func(t *testing.T) {
//			t.Parallel()  // REQUIRED - FIRST LINE
//			testSMTP := GetTestSMTP(t)  // REQUIRED - own instance
//			client := testSMTP.MailpitClient
//			ctx := t.Context()
//
//			// Send test email specific to this test
//			sendTestEmailWithSubject(t, testSMTP, "GetMessage Test Email")
//			messages := testSMTP.WaitForMessages(t, 1, 5*time.Second)
//
//			message, err := client.GetMessage(ctx, messages[0].ID)
//			require.NoError(t, err)
//			require.Equal(t, "GetMessage Test Email", message.Subject)
//		})
//	}
//
// # Container Management
//
// The package uses a container pool to optimize test performance while maintaining isolation:
//
// - Container Pool: Reuses Mailpit containers across tests for efficiency
// - Automatic Cleanup: Containers are properly cleaned up after test completion
// - TLS Support: Containers are pre-configured with TLS certificates for HTTPS testing
// - Port Management: Dynamically assigned ports prevent conflicts
//
// Container pool size can be configured via environment variable:
//
//	export TEST_SMTP_POOL_SIZE=10  # Default is 5
//
// # TestSMTP Structure
//
// The TestSMTP struct provides everything needed for e2e testing:
//
//	type TestSMTP struct {
//		Container     testcontainers.Container  // Docker container instance
//		MailpitClient mailpitclient.Client     // Pre-configured mailpit client
//		SMTPPort      string                   // Mapped SMTP port (1025)
//		APIPort       string                   // Mapped API port (8025)
//		Host          string                   // Container host (usually localhost)
//		SMTPConfig    SMTPConfig              // SMTP connection details
//	}
//
// # Helper Functions
//
// The package provides several helper functions for common testing scenarios:
//
// ## ClearMessages
// Removes all messages from the Mailpit server:
//
//	testSMTP.ClearMessages(t)
//
// ## GetMessages
// Retrieves all current messages:
//
//	messages := testSMTP.GetMessages(t)
//
// ## WaitForMessages
// Waits for a specific number of messages with timeout:
//
//	messages := testSMTP.WaitForMessages(t, 2, 10*time.Second)
//
// # SMTP Configuration
//
// The SMTPConfig provides SMTP server connection details:
//
//	smtpConfig := testSMTP.SMTPConfig
//	// Connect to SMTP server at smtpConfig.Host:smtpConfig.Port
//	// Supports both TLS and non-TLS connections
//	// Authentication: PLAIN (accepts any credentials)
//
// Example SMTP usage with Go's net/smtp:
//
//	func sendTestEmail(t *testing.T, testSMTP *TestSMTP, subject, body string) {
//		config := testSMTP.SMTPConfig
//		addr := net.JoinHostPort(config.Host, strconv.Itoa(int(config.Port)))
//
//		msg := fmt.Sprintf("To: test@example.com\r\nSubject: %s\r\n\r\n%s", subject, body)
//		err := smtp.SendMail(addr, nil, "sender@example.com", []string{"test@example.com"}, []byte(msg))
//		require.NoError(t, err)
//	}
//
// # Complete E2E Test Example
//
// Here's a comprehensive example showing proper e2e test structure:
//
//	func TestMailpitClient_E2E(t *testing.T) {
//		t.Parallel()
//
//		t.Run("SendAndRetrieve", func(t *testing.T) {
//			t.Parallel()
//			testSMTP := GetTestSMTP(t)
//			client := testSMTP.MailpitClient
//			ctx := t.Context()
//
//			// Clear any existing messages
//			testSMTP.ClearMessages(t)
//
//			// Send a test email via SMTP
//			sendTestEmail(t, testSMTP, "E2E Test Subject", "Test body content")
//
//			// Wait for message to arrive
//			messages := testSMTP.WaitForMessages(t, 1, 10*time.Second)
//			require.Len(t, messages, 1)
//
//			// Verify message content
//			message, err := client.GetMessage(ctx, messages[0].ID)
//			require.NoError(t, err)
//			require.Equal(t, "E2E Test Subject", message.Subject)
//
//			// Test message operations
//			err = client.MarkMessageRead(ctx, message.ID)
//			require.NoError(t, err)
//
//			// Get updated message
//			updatedMessage, err := client.GetMessage(ctx, message.ID)
//			require.NoError(t, err)
//			require.True(t, updatedMessage.Read)
//
//			// Clean up
//			err = client.DeleteMessage(ctx, message.ID)
//			require.NoError(t, err)
//		})
//	}
//
// # Performance Considerations
//
// - Container Pooling: Containers are reused across tests to reduce startup overhead
// - Parallel Execution: All tests run in parallel for maximum performance
// - Resource Cleanup: Automatic cleanup prevents resource leaks
// - Wait Strategies: Use WaitForMessages() instead of fixed sleeps when possible
//
// # TestMain Setup
//
// For proper cleanup, use TestMain in your test files:
//
//	func TestMain(m *testing.M) {
//		code := m.Run()
//		testing.CleanupSMTPContainers()
//		os.Exit(code)
//	}
//
// # Common Pitfalls to Avoid
//
// 1. DON'T share TestSMTP instances between tests - causes race conditions
// 2. DON'T forget t.Parallel() - breaks parallel execution requirements
// 3. DON'T use fixed sleeps - use WaitForMessages() for reliability
// 4. DON'T rely on message order - tests should be order-independent
// 5. DON'T assume clean state - use ClearMessages() or send specific test data
//
// # Environment Configuration
//
// The testing package supports these environment variables:
//
//	TEST_SMTP_POOL_SIZE=5    # Container pool size (default: 5)
//
// # Dependencies
//
// Required for e2e testing:
// - Docker daemon running
// - testcontainers-go library
// - Mailpit Docker image (axllent/mailpit:latest)
// - TLS certificates in certs/ directory (generated via make mkcert-generate)
//
// This testing framework ensures reliable, fast, and isolated e2e testing for the
// mailpitclient library while maintaining production-level quality standards.
package testing

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/CodeLieutenant/mailpitclient"
)

// SMTPContainerPool manages a pool of SMTP containers
type SMTPContainerPool struct {
	available  chan testcontainers.Container
	containers []testcontainers.Container
	maxSize    int
	created    int
	mu         sync.RWMutex
}

var (
	smtpContainerPool *SMTPContainerPool
	smtpPoolMu        sync.Mutex
)

type SMTPConfig struct {
	Host       string
	Username   string
	Password   string
	AuthType   string
	Encryption string
	Port       uint16
}

// TestSMTP holds the test SMTP resources
type TestSMTP struct {
	Container     testcontainers.Container
	MailpitClient mailpitclient.Client
	SMTPPort      string
	APIPort       string
	Host          string
	SMTPConfig    SMTPConfig
}

type TestSMTPOptions struct {
	MailPitClientConfig *mailpitclient.Config
	SMTPConfig          *SMTPConfig
	MailpitImage        string
	MailpitEnv          map[string]string
	MailpitKey          string
	MailpitCert         string
}

type Option func(*TestSMTPOptions)

func WithMailPitClientConfig(mailPitClientConfig *mailpitclient.Config) Option {
	return func(opts *TestSMTPOptions) {
		opts.MailPitClientConfig = mailPitClientConfig
	}
}

func WithSMTPConfig(smtpConfig *SMTPConfig) Option {
	return func(opts *TestSMTPOptions) {
		opts.SMTPConfig = smtpConfig
	}
}

func WithMailPitImage(mailPitImage string) Option {
	return func(opts *TestSMTPOptions) {
		opts.MailpitImage = mailPitImage
	}
}

func WithMailPitEnv(mailPitEnv map[string]string) Option {
	return func(opts *TestSMTPOptions) {
		opts.MailpitEnv = mailPitEnv
	}
}

func WithMailPitKey(mailPitKey string) Option {
	return func(opts *TestSMTPOptions) {
		opts.MailpitKey = mailPitKey
	}
}

func WithMailPitCert(mailPitCert string) Option {
	return func(opts *TestSMTPOptions) {
		opts.MailpitCert = mailPitCert
	}
}

// GetTestSMTP returns a configured SMTP test environment with mailpit container.
// It uses a singleton container for efficiency and proper resource management.
func GetTestSMTP(tb testing.TB, opts ...Option) *TestSMTP {
	tb.Helper()

	ctx := tb.Context()

	testOpts := TestSMTPOptions{
		MailPitClientConfig: &mailpitclient.Config{
			APIPath:    "/api/v1",
			HTTPClient: &http.Client{Timeout: 10 * time.Second},
		},
		SMTPConfig: &SMTPConfig{
			AuthType:   "PLAIN",
			Encryption: "starttls",
		},
	}

	for _, opt := range opts {
		opt(&testOpts)
	}

	// Use pooled container for parallel testing support
	container := getSMTPContainerFromPool(tb,
		testOpts.MailpitImage,
		testOpts.MailpitEnv,
		testOpts.MailpitKey,
		testOpts.MailpitCert,
	)
	tb.Cleanup(func() {
		releaseSMTPContainerToPool(container)
	})

	// Get the mapped ports
	smtpPort, err := container.MappedPort(ctx, "1025")
	if err != nil {
		tb.Fatalf("Failed to get SMTP port: %v", err)
	}

	apiPort, err := container.MappedPort(ctx, "8025")
	if err != nil {
		tb.Fatalf("Failed to get API port: %v", err)
	}

	// Get the container host
	host, err := container.Host(ctx)
	if err != nil {
		tb.Fatalf("Failed to get container host: %v", err)
	}

	testOpts.MailPitClientConfig.BaseURL = "http://" + net.JoinHostPort(host, apiPort.Port())
	testOpts.SMTPConfig.Host = host
	testOpts.SMTPConfig.Port = uint16(smtpPort.Int())

	mailpitClient, err := mailpitclient.NewClient(testOpts.MailPitClientConfig)
	if err != nil {
		tb.Fatalf("Failed to create mailpit client: %v", err)
	}

	// Setup per-test cleanup
	tb.Cleanup(func() {
		if mailpitClient != nil {
			if err = mailpitClient.Close(); err != nil {
				tb.Errorf("Failed to close mailpit client: %v", err)
			}
		}
	})

	return &TestSMTP{
		Container:     container,
		SMTPConfig:    *testOpts.SMTPConfig,
		MailpitClient: mailpitClient,
		SMTPPort:      smtpPort.Port(),
		APIPort:       apiPort.Port(),
		Host:          host,
	}
}

// initSMTPContainerPool initializes the SMTP container pool structure (lazy creation)
func initSMTPContainerPool(tb testing.TB) {
	tb.Helper()

	smtpPoolMu.Lock()
	defer smtpPoolMu.Unlock()

	if smtpContainerPool != nil {
		return
	}

	poolSize := 5
	// Fixed pool size for SMTP containers (can be configurable)
	if envPoolSize := os.Getenv("TEST_SMTP_POOL_SIZE"); envPoolSize != "" {
		if size, err := strconv.Atoi(envPoolSize); err == nil && size > 0 {
			poolSize = size
		}
	}

	smtpContainerPool = &SMTPContainerPool{
		containers: make([]testcontainers.Container, 0, poolSize),
		available:  make(chan testcontainers.Container, poolSize),
		maxSize:    poolSize,
		created:    0,
	}
}

// getSMTPContainerFromPool gets a container from the pool, creating one lazily if needed
func getSMTPContainerFromPool(tb testing.TB, image string, envs map[string]string, keyPath, crtPath string) testcontainers.Container {
	tb.Helper()

	initSMTPContainerPool(tb)

	// Try to get an available container first (non-blocking)
	select {
	case c := <-smtpContainerPool.available:
		return c
	default:
		// No available containers, try to create one if we haven't reached the limit
	}

	// Check if we can create a new container (within bounds)
	smtpContainerPool.mu.Lock()
	canCreate := smtpContainerPool.created < smtpContainerPool.maxSize
	if canCreate {
		smtpContainerPool.created++
	}
	smtpContainerPool.mu.Unlock()

	//nolint:nestif
	if canCreate {
		// Create a new container lazily
		ctx := tb.Context()

		// Get project root and certificates directory
		certsPath := filepath.Join(projectRootDir(tb), "certs")

		if image == "" {
			image = "axllent/mailpit:latest"
		}

		defaultEnv := map[string]string{
			"MP_SMTP_REQUIRE_STARTTLS":    "false", // Allow both TLS and non-TLS connections
			"MP_ENABLE_SPAMASSASSIN":      "true",
			"MP_SMTP_AUTH_ACCEPT_ANY":     "1",
			"MP_SMTP_AUTH_ALLOW_INSECURE": "1",
			"MP_SMTP_8BITMIME":            "1", // Enable 8BITMIME support
		}

		for k, v := range envs {
			defaultEnv[k] = v
		}

		if keyPath == "" {
			keyPath = filepath.Join(certsPath, "smtp.key")
		}

		if crtPath == "" {
			crtPath = filepath.Join(certsPath, "smtp.crt")
		}

		files := make([]testcontainers.ContainerFile, 0, 2)
		if _, err := os.Stat(crtPath); err == nil {
			files = append(files, testcontainers.ContainerFile{
				HostFilePath:      crtPath,
				ContainerFilePath: "/certs/smtp.crt",
			})

			defaultEnv["MP_SMTP_TLS_CERT"] = "/certs/smtp.crt"
		}

		if _, err := os.Stat(keyPath); err == nil {
			files = append(files, testcontainers.ContainerFile{
				HostFilePath:      keyPath,
				ContainerFilePath: "/certs/smtp.key",
			})

			defaultEnv["MP_SMTP_TLS_KEY"] = "/certs/smtp.key"
		}

		// Create mailpit container request
		req := testcontainers.ContainerRequest{
			Image:        image,
			ExposedPorts: []string{"1025/tcp", "8025/tcp"},
			WaitingFor: wait.ForAll(
				wait.ForListeningPort("1025/tcp"),
				wait.ForListeningPort("8025/tcp"),
				wait.ForHTTP("/api/v1/info").WithPort("8025/tcp").WithStartupTimeout(30*time.Second),
			),
			Env:   defaultEnv,
			Files: files,
		}

		// Start the container
		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			// Decrement counter on failure
			smtpContainerPool.mu.Lock()
			smtpContainerPool.created--
			smtpContainerPool.mu.Unlock()
			tb.Fatalf("Failed to start mailpit container: %v", err)
		}

		smtpContainerPool.mu.Lock()
		smtpContainerPool.containers = append(smtpContainerPool.containers, container)
		smtpContainerPool.mu.Unlock()

		return container
	}

	// Wait for an available container (blocking)
	select {
	case cont := <-smtpContainerPool.available:
		return cont
	case <-tb.Context().Done():
		tb.Fatalf("Test context cancelled while waiting for SMTP container: %v", tb.Context().Err())
	}

	return nil
}

// releaseSMTPContainerToPool returns a container to the pool
func releaseSMTPContainerToPool(container testcontainers.Container) {
	smtpContainerPool.available <- container
}

// ClearMessages is a helper function to clear all messages from mailpit
func (ts *TestSMTP) ClearMessages(tb testing.TB) {
	tb.Helper()
	err := ts.MailpitClient.DeleteAllMessages(tb.Context())
	if err != nil {
		tb.Fatalf("Failed to clear messages: %v", err)
	}
}

// GetMessages is a helper function to retrieve all messages from mailpit
func (ts *TestSMTP) GetMessages(tb testing.TB) []mailpitclient.Message {
	tb.Helper()
	resp, err := ts.MailpitClient.ListMessages(tb.Context(), nil)
	if err != nil {
		tb.Fatalf("Failed to get messages from mailpit API: %v", err)
	}

	return resp.Messages
}

// WaitForMessages waits for the expected number of messages to arrive
func (ts *TestSMTP) WaitForMessages(tb testing.TB, expectedCount int, timeout time.Duration) []mailpitclient.Message {
	tb.Helper()

	ctx, cancel := context.WithTimeout(tb.Context(), timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			tb.Fatalf("Timeout waiting for %d messages", expectedCount)

			return nil
		case <-ticker.C:
			messages := ts.GetMessages(tb)
			if len(messages) >= expectedCount {
				return messages
			}
		}
	}
}

func CleanupSMTPContainers() {
	smtpPoolMu.Lock()
	defer smtpPoolMu.Unlock()

	if smtpContainerPool == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Close the available channel to prevent new acquisitions
	close(smtpContainerPool.available)

	// Terminate all containers
	for _, c := range smtpContainerPool.containers {
		go func(container testcontainers.Container) {
			if err := container.Terminate(ctx); err != nil {
				log.Printf("Failed to terminate container: %v", err)
			}
		}(c)
	}

	smtpContainerPool = nil
}

const gomod = "go.mod"

var (
	projectRootDirCache   = make(map[string]string, 10)
	projectRootDirCacheMu sync.RWMutex
)

func projectRootDir(tb testing.TB) string {
	tb.Helper()
	originalWorkingDir := workingDir(tb)
	wd := originalWorkingDir

	projectRootDirCacheMu.RLock()
	if dir, ok := projectRootDirCache[originalWorkingDir]; ok {
		projectRootDirCacheMu.RUnlock()

		return dir
	}
	projectRootDirCacheMu.RUnlock()

	for entries, err := os.ReadDir(wd); err == nil; {
		for _, entry := range entries {
			if entry.Name() == gomod {
				projectRootDirCacheMu.Lock()
				projectRootDirCache[originalWorkingDir] = wd
				projectRootDirCacheMu.Unlock()

				return wd
			}
		}

		if wd == "/" {
			tb.Error("got to FS Root, file not found")
			tb.FailNow()
		}

		wd, err = getAbsolutePath(filepath.Join(wd, ".."))
		if err != nil {
			tb.Errorf("failed to get absolute path from %s", filepath.Join(wd, ".."))
			tb.FailNow()
		}

		entries, err = os.ReadDir(wd)
	}

	tb.Errorf("%s not found", gomod)
	tb.FailNow()

	return ""
}

func workingDir(tb testing.TB) string {
	tb.Helper()
	wd, err := os.Getwd()
	if err != nil {
		tb.Error(err)
		tb.FailNow()
	}

	return wd
}

// getAbsolutePath Returns absolute path string for a given directory and error if directory doesent exist
func getAbsolutePath(path string) (string, error) {
	var err error

	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return "", err
		}

		return path, nil
	}

	return path, err
}
