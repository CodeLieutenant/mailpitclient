package mailpit_go_api

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
	Port       uint16
	Username   string
	Password   string
	AuthType   string
	Encryption string
}

// TestSMTP holds the test SMTP resources
type TestSMTP struct {
	Container     testcontainers.Container
	MailpitClient Client
	SMTPPort      string
	APIPort       string
	Host          string
	SMTPConfig    SMTPConfig
}

// GetTestSMTP returns a configured SMTP test environment with mailpit container.
// It uses a singleton container for efficiency and proper resource management.
func GetTestSMTP(tb testing.TB) *TestSMTP {
	tb.Helper()

	ctx := tb.Context()

	// Use pooled container for parallel testing support
	container := getSMTPContainerFromPool(tb)
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

	// Create SMTP service configuration
	smtpConfig := SMTPConfig{
		Host:       host,
		Port:       uint16(smtpPort.Int()),
		Username:   "",
		Password:   "",
		AuthType:   "PLAIN",
		Encryption: "starttls",
	}
	// Create mailpit client
	mailpitConfig := &Config{
		BaseURL:    "http://" + net.JoinHostPort(host, apiPort.Port()),
		APIPath:    "/api/v1",
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}

	mailpitClient, err := NewClient(mailpitConfig)
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
		SMTPConfig:    smtpConfig,
		MailpitClient: mailpitClient,
		SMTPPort:      smtpPort.Port(),
		APIPort:       apiPort.Port(),
		Host:          host,
	}
}

// initSMTPContainerPool initializes the SMTP container pool structure (lazy creation)
func initSMTPContainerPool(tb testing.TB) {
	tb.Helper()

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
func getSMTPContainerFromPool(tb testing.TB) testcontainers.Container {
	tb.Helper()

	smtpPoolMu.Lock()
	defer smtpPoolMu.Unlock()

	if smtpContainerPool == nil {
		initSMTPContainerPool(tb)
	}

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

	if canCreate {
		// Create a new container lazily
		ctx := tb.Context()

		// Get project root and certificates directory
		certsPath := filepath.Join(ProjectRootDir(tb), "certs")

		// Create mailpit container request
		req := testcontainers.ContainerRequest{
			Image:        "axllent/mailpit:latest",
			ExposedPorts: []string{"1025/tcp", "8025/tcp"},
			WaitingFor: wait.ForAll(
				wait.ForListeningPort("1025/tcp"),
				wait.ForListeningPort("8025/tcp"),
				wait.ForHTTP("/api/v1/info").WithPort("8025/tcp").WithStartupTimeout(30*time.Second),
			),
			Env: map[string]string{
				"MP_SMTP_TLS_CERT":            "/certs/smtp.crt",
				"MP_SMTP_TLS_KEY":             "/certs/smtp.key",
				"MP_SMTP_REQUIRE_STARTTLS":    "false", // Allow both TLS and non-TLS connections
				"MP_ENABLE_SPAMASSASSIN":      "true",
				"MP_SMTP_AUTH_ACCEPT_ANY":     "1",
				"MP_SMTP_AUTH_ALLOW_INSECURE": "1",
				"MP_SMTP_8BITMIME":            "1", // Enable 8BITMIME support
			},
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      filepath.Join(certsPath, "smtp.crt"),
					ContainerFilePath: "/certs/smtp.crt",
				},
				{
					HostFilePath:      filepath.Join(certsPath, "smtp.key"),
					ContainerFilePath: "/certs/smtp.key",
				},
			},
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
	if smtpContainerPool != nil {
		smtpContainerPool.available <- container
	}
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
func (ts *TestSMTP) GetMessages(tb testing.TB) []Message {
	tb.Helper()
	resp, err := ts.MailpitClient.ListMessages(tb.Context(), nil)
	if err != nil {
		tb.Fatalf("Failed to get messages from mailpit API: %v", err)
	}

	return resp.Messages
}

// WaitForMessages waits for the expected number of messages to arrive
func (ts *TestSMTP) WaitForMessages(tb testing.TB, expectedCount int, timeout time.Duration) []Message {
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

// cleanupSMTPContainerPool terminates all containers in the pool
func cleanupSMTPContainerPool() {
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

// CleanupSMTPContainers should be called in TestMain to cleanup shared SMTP resources
func CleanupSMTPContainers() {
	smtpPoolMu.Lock()
	defer smtpPoolMu.Unlock()
	cleanupSMTPContainerPool()
}

const gomod = "go.mod"

var (
	projectRootDirCache   = make(map[string]string, 10)
	projectRootDirCacheMu sync.RWMutex
)

func ProjectRootDir(tb testing.TB) string {
	tb.Helper()
	originalWorkingDir := WorkingDir(tb)
	workingDir := originalWorkingDir

	projectRootDirCacheMu.RLock()
	if dir, ok := projectRootDirCache[originalWorkingDir]; ok {
		projectRootDirCacheMu.RUnlock()

		return dir
	}
	projectRootDirCacheMu.RUnlock()

	for entries, err := os.ReadDir(workingDir); err == nil; {
		for _, entry := range entries {
			if entry.Name() == gomod {
				projectRootDirCacheMu.Lock()
				projectRootDirCache[originalWorkingDir] = workingDir
				projectRootDirCacheMu.Unlock()

				return workingDir
			}
		}

		if workingDir == "/" {
			tb.Error("got to FS Root, file not found")
			tb.FailNow()
		}

		workingDir, err = GetAbsolutePath(filepath.Join(workingDir, ".."))
		if err != nil {
			tb.Errorf("failed to get absolute path from %s", filepath.Join(workingDir, ".."))
			tb.FailNow()
		}

		entries, err = os.ReadDir(workingDir)
	}

	tb.Errorf("%s not found", gomod)
	tb.FailNow()

	return ""
}

func WorkingDir(tb testing.TB) string {
	tb.Helper()
	wd, err := os.Getwd()
	if err != nil {
		tb.Error(err)
		tb.FailNow()
	}

	return wd
}

// GetAbsolutePath Returns absolute path string for a given directory and error if directory doesent exist
func GetAbsolutePath(path string) (string, error) {
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
