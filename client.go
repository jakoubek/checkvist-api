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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
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

// authResponse represents the response from the authentication endpoint.
type authResponse struct {
	Token string `json:"token"`
}

// Authenticate performs explicit authentication with the Checkvist API.
// This is optional - the client will automatically authenticate when needed.
// Use this to verify credentials or to pre-authenticate before making requests.
func (c *Client) Authenticate(ctx context.Context) error {
	return c.authenticate(ctx, "")
}

// AuthenticateWith2FA performs authentication with a 2FA token.
func (c *Client) AuthenticateWith2FA(ctx context.Context, twoFAToken string) error {
	return c.authenticate(ctx, twoFAToken)
}

// authenticate performs the actual authentication request.
func (c *Client) authenticate(ctx context.Context, twoFAToken string) error {
	data := url.Values{}
	data.Set("username", c.username)
	data.Set("remote_key", c.remoteKey)
	if twoFAToken != "" {
		data.Set("totp", twoFAToken)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/auth/login.json?version=2",
		strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return NewAPIError(resp, string(body))
	}

	var authResp authResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("decoding auth response: %w", err)
	}

	c.mu.Lock()
	c.token = authResp.Token
	// Token is valid for 1 day, but we refresh earlier to be safe
	c.tokenExp = time.Now().Add(23 * time.Hour)
	c.mu.Unlock()

	c.logger.Debug("authenticated successfully", "username", c.username)
	return nil
}

// refreshToken renews the authentication token.
func (c *Client) refreshToken(ctx context.Context) error {
	c.mu.RLock()
	currentToken := c.token
	c.mu.RUnlock()

	if currentToken == "" {
		return c.Authenticate(ctx)
	}

	data := url.Values{}
	data.Set("old_token", currentToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/auth/refresh_token.json?version=2",
		strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// If refresh fails, try full authentication
		c.logger.Debug("token refresh failed, attempting full authentication")
		return c.Authenticate(ctx)
	}

	var authResp authResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("decoding refresh response: %w", err)
	}

	c.mu.Lock()
	c.token = authResp.Token
	// Refreshed tokens can be valid for up to 90 days, but we refresh more frequently
	c.tokenExp = time.Now().Add(23 * time.Hour)
	c.mu.Unlock()

	c.logger.Debug("token refreshed successfully")
	return nil
}

// ensureAuthenticated ensures the client has a valid authentication token.
// This is called automatically before each API request.
func (c *Client) ensureAuthenticated(ctx context.Context) error {
	c.mu.RLock()
	token := c.token
	tokenExp := c.tokenExp
	c.mu.RUnlock()

	if token == "" {
		return c.Authenticate(ctx)
	}

	// Refresh token if it will expire within the next hour
	if time.Now().Add(1 * time.Hour).After(tokenExp) {
		return c.refreshToken(ctx)
	}

	return nil
}

// getToken returns the current authentication token.
// Thread-safe.
func (c *Client) getToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token
}

// CurrentUser returns information about the currently authenticated user.
func (c *Client) CurrentUser(ctx context.Context) (*User, error) {
	if err := c.ensureAuthenticated(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/auth/curr_user.json", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-Client-Token", c.getToken())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, NewAPIError(resp, string(body))
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &user, nil
}
