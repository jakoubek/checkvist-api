// Package checkvist provides a type-safe, idiomatic Go client for the Checkvist API.
//
// This package allows Go applications to interact with Checkvist checklists,
// tasks, and notes. It handles authentication, automatic token renewal,
// and provides fluent interfaces for task creation and filtering.
//
// Basic usage:
//
//	client := checkvist.NewClient(username, remoteKey)
//	checklists, err := client.Checklists().List(ctx)
package checkvist

import (
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// client.go contains the Client struct, constructor, and authentication logic.

const (
	// DefaultBaseURL is the default base URL for the Checkvist API.
	DefaultBaseURL = "https://checkvist.com"
	// DefaultTimeout is the default timeout for HTTP requests.
	DefaultTimeout = 30 * time.Second
)

// Client is the Checkvist API client.
type Client struct {
	// baseURL is the base URL for API requests.
	baseURL string
	// username is the user's email address.
	username string
	// remoteKey is the API key (remote key) for authentication.
	remoteKey string
	// token is the current authentication token.
	token string
	// tokenExp is the expiration time of the current token.
	tokenExp time.Time
	// httpClient is the HTTP client used for requests.
	httpClient *http.Client
	// retryConf is the retry configuration for failed requests.
	retryConf RetryConfig
	// logger is the logger for debug and error messages.
	logger *slog.Logger
	// mu protects token and tokenExp for concurrent access.
	mu sync.RWMutex
}

// NewClient creates a new Checkvist API client.
//
// The username should be the user's email address, and remoteKey is the API key
// which can be obtained from Checkvist settings.
//
// Example:
//
//	client := checkvist.NewClient("user@example.com", "your-api-key")
//
// With options:
//
//	client := checkvist.NewClient("user@example.com", "your-api-key",
//	    checkvist.WithTimeout(60 * time.Second),
//	    checkvist.WithRetryConfig(checkvist.RetryConfig{MaxRetries: 5}),
//	)
func NewClient(username, remoteKey string, opts ...Option) *Client {
	c := &Client{
		baseURL:   DefaultBaseURL,
		username:  username,
		remoteKey: remoteKey,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		retryConf: DefaultRetryConfig(),
		logger:    slog.Default(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
