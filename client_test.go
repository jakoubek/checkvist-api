package checkvist

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

// loadFixture loads a JSON fixture file from testdata.
func loadFixture(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to load fixture %s: %v", path, err)
	}
	return data
}

func TestNewClient_Defaults(t *testing.T) {
	client := NewClient("user@example.com", "remote-key")

	if client.baseURL != DefaultBaseURL {
		t.Errorf("expected baseURL %s, got %s", DefaultBaseURL, client.baseURL)
	}
	if client.username != "user@example.com" {
		t.Errorf("expected username user@example.com, got %s", client.username)
	}
	if client.remoteKey != "remote-key" {
		t.Errorf("expected remoteKey remote-key, got %s", client.remoteKey)
	}
	if client.httpClient == nil {
		t.Error("expected httpClient to be set")
	}
	if client.retryConf.MaxRetries != 3 {
		t.Errorf("expected MaxRetries 3, got %d", client.retryConf.MaxRetries)
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	customClient := &http.Client{Timeout: 60 * time.Second}
	customLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	customRetry := RetryConfig{MaxRetries: 5, BaseDelay: 2 * time.Second}

	client := NewClient("user@example.com", "remote-key",
		WithHTTPClient(customClient),
		WithLogger(customLogger),
		WithRetryConfig(customRetry),
		WithBaseURL("https://custom.api.com"),
	)

	if client.httpClient != customClient {
		t.Error("expected custom HTTP client")
	}
	if client.logger != customLogger {
		t.Error("expected custom logger")
	}
	if client.retryConf.MaxRetries != 5 {
		t.Errorf("expected MaxRetries 5, got %d", client.retryConf.MaxRetries)
	}
	if client.baseURL != "https://custom.api.com" {
		t.Errorf("expected custom baseURL, got %s", client.baseURL)
	}
}

func TestAuthenticate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/login.json" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// Verify form data
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}
		if r.Form.Get("username") != "user@example.com" {
			t.Errorf("expected username user@example.com, got %s", r.Form.Get("username"))
		}
		if r.Form.Get("remote_key") != "api-key" {
			t.Errorf("expected remote_key api-key, got %s", r.Form.Get("remote_key"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(loadFixture(t, "testdata/auth/login_success.json"))
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	err := client.Authenticate(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.token != "test-token-abc123" {
		t.Errorf("expected token test-token-abc123, got %s", client.token)
	}
}

func TestAuthenticate_InvalidCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Invalid credentials"}`))
	}))
	defer server.Close()

	client := NewClient("user@example.com", "wrong-key", WithBaseURL(server.URL))
	err := client.Authenticate(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Error("expected error to wrap ErrUnauthorized")
	}
}

func TestAuthenticate_2FA(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}

		// Verify 2FA token is sent
		if r.Form.Get("totp") != "123456" {
			t.Errorf("expected totp 123456, got %s", r.Form.Get("totp"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(loadFixture(t, "testdata/auth/login_success.json"))
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	err := client.AuthenticateWith2FA(context.Background(), "123456")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTokenRefresh_Auto(t *testing.T) {
	var authCalls int32
	var refreshCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			atomic.AddInt32(&authCalls, 1)
			json.NewEncoder(w).Encode(map[string]string{"token": "initial-token"})
		case "/auth/refresh_token.json":
			atomic.AddInt32(&refreshCalls, 1)
			json.NewEncoder(w).Encode(map[string]string{"token": "refreshed-token"})
		case "/test":
			// Verify token is sent
			if r.Header.Get("X-Client-Token") == "" {
				t.Error("expected X-Client-Token header")
			}
			w.Write([]byte(`{"ok": true}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))

	// First call should authenticate
	err := client.ensureAuthenticated(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&authCalls) != 1 {
		t.Errorf("expected 1 auth call, got %d", authCalls)
	}

	// Simulate token about to expire
	client.mu.Lock()
	client.tokenExp = time.Now().Add(30 * time.Minute) // Less than 1 hour
	client.mu.Unlock()

	// This should trigger a refresh
	err = client.ensureAuthenticated(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&refreshCalls) != 1 {
		t.Errorf("expected 1 refresh call, got %d", refreshCalls)
	}
}

func TestTokenRefresh_Manual(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "initial-token"})
		case "/auth/refresh_token.json":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}
			if r.Form.Get("old_token") != "initial-token" {
				t.Errorf("expected old_token initial-token, got %s", r.Form.Get("old_token"))
			}
			json.NewEncoder(w).Encode(map[string]string{"token": "refreshed-token"})
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))

	// First authenticate
	err := client.Authenticate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.token != "initial-token" {
		t.Errorf("expected initial-token, got %s", client.token)
	}

	// Manually refresh
	err = client.refreshToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.token != "refreshed-token" {
		t.Errorf("expected refreshed-token, got %s", client.token)
	}
}

func TestCurrentUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/auth/curr_user.json":
			if r.Header.Get("X-Client-Token") != "test-token" {
				t.Errorf("expected X-Client-Token test-token, got %s", r.Header.Get("X-Client-Token"))
			}
			w.Write(loadFixture(t, "testdata/auth/current_user.json"))
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key", WithBaseURL(server.URL))
	user, err := client.CurrentUser(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != 12345 {
		t.Errorf("expected ID 12345, got %d", user.ID)
	}
	if user.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", user.Username)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
}

func TestRetryLogic_429(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/test":
			count := atomic.AddInt32(&attempts, 1)
			if count < 3 {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error": "rate limited"}`))
				return
			}
			w.Write([]byte(`{"success": true}`))
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key",
		WithBaseURL(server.URL),
		WithRetryConfig(RetryConfig{
			MaxRetries: 5,
			BaseDelay:  1 * time.Millisecond,
			MaxDelay:   10 * time.Millisecond,
			Jitter:     false,
		}),
	)

	var result map[string]bool
	err := client.doGet(context.Background(), "/test", &result)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
	if !result["success"] {
		t.Error("expected success=true in response")
	}
}

func TestRetryLogic_5xx(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/test":
			count := atomic.AddInt32(&attempts, 1)
			if count < 2 {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "server error"}`))
				return
			}
			w.Write([]byte(`{"success": true}`))
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key",
		WithBaseURL(server.URL),
		WithRetryConfig(RetryConfig{
			MaxRetries: 3,
			BaseDelay:  1 * time.Millisecond,
			MaxDelay:   10 * time.Millisecond,
			Jitter:     false,
		}),
	)

	var result map[string]bool
	err := client.doGet(context.Background(), "/test", &result)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestRetryLogic_ExhaustedRetries(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/test":
			atomic.AddInt32(&attempts, 1)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error": "service unavailable"}`))
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key",
		WithBaseURL(server.URL),
		WithRetryConfig(RetryConfig{
			MaxRetries: 2,
			BaseDelay:  1 * time.Millisecond,
			MaxDelay:   10 * time.Millisecond,
			Jitter:     false,
		}),
	)

	var result map[string]bool
	err := client.doGet(context.Background(), "/test", &result)

	if err == nil {
		t.Fatal("expected error after exhausted retries")
	}
	// 1 initial + 2 retries = 3 total attempts
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryLogic_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/login.json":
			json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
		case "/test":
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "rate limited"}`))
		}
	}))
	defer server.Close()

	client := NewClient("user@example.com", "api-key",
		WithBaseURL(server.URL),
		WithRetryConfig(RetryConfig{
			MaxRetries: 10,
			BaseDelay:  100 * time.Millisecond,
			MaxDelay:   1 * time.Second,
			Jitter:     false,
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	var result map[string]bool
	err := client.doGet(ctx, "/test", &result)

	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestCalculateRetryDelay(t *testing.T) {
	client := NewClient("user", "key",
		WithRetryConfig(RetryConfig{
			MaxRetries: 5,
			BaseDelay:  100 * time.Millisecond,
			MaxDelay:   1 * time.Second,
			Jitter:     false,
		}),
	)

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 200 * time.Millisecond}, // 100ms * 2^1
		{2, 400 * time.Millisecond}, // 100ms * 2^2
		{3, 800 * time.Millisecond}, // 100ms * 2^3
		{4, 1 * time.Second},        // capped at MaxDelay
		{5, 1 * time.Second},        // capped at MaxDelay
	}

	for _, tc := range tests {
		delay := client.calculateRetryDelay(tc.attempt)
		if delay != tc.expected {
			t.Errorf("attempt %d: expected %v, got %v", tc.attempt, tc.expected, delay)
		}
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("expected MaxRetries 3, got %d", config.MaxRetries)
	}
	if config.BaseDelay != 1*time.Second {
		t.Errorf("expected BaseDelay 1s, got %v", config.BaseDelay)
	}
	if config.MaxDelay != 30*time.Second {
		t.Errorf("expected MaxDelay 30s, got %v", config.MaxDelay)
	}
	if !config.Jitter {
		t.Error("expected Jitter to be true")
	}
}
