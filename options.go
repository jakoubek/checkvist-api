package checkvist

import (
	"log/slog"
	"net/http"
	"time"
)

// options.go contains functional options for configuring the Client.

// RetryConfig configures the retry behavior for failed requests.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts.
	MaxRetries int
	// BaseDelay is the initial delay before the first retry.
	BaseDelay time.Duration
	// MaxDelay is the maximum delay between retries.
	MaxDelay time.Duration
	// Jitter enables randomized delay to prevent thundering herd.
	Jitter bool
}

// DefaultRetryConfig returns the default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
		Jitter:     true,
	}
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client for the Checkvist client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithTimeout sets the timeout for HTTP requests.
// This creates a new HTTP client with the specified timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient = &http.Client{
			Timeout: timeout,
		}
	}
}

// WithRetryConfig sets the retry configuration for failed requests.
func WithRetryConfig(config RetryConfig) Option {
	return func(c *Client) {
		c.retryConf = config
	}
}

// WithLogger sets a custom logger for the client.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithBaseURL sets a custom base URL for the API.
// This is primarily useful for testing.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}
