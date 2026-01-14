package checkvist

import (
	"errors"
	"fmt"
	"net/http"
)

// errors.go contains the APIError type and sentinel errors for common API error conditions.

// Sentinel errors for common API error conditions.
// Use errors.Is() to check for these errors.
var (
	// ErrUnauthorized is returned when authentication fails (HTTP 401).
	ErrUnauthorized = errors.New("unauthorized: invalid credentials or expired token")
	// ErrNotFound is returned when a resource is not found (HTTP 404).
	ErrNotFound = errors.New("not found: the requested resource does not exist")
	// ErrRateLimited is returned when the API rate limit is exceeded (HTTP 429).
	ErrRateLimited = errors.New("rate limited: too many requests")
	// ErrBadRequest is returned for invalid request parameters (HTTP 400).
	ErrBadRequest = errors.New("bad request: invalid parameters")
	// ErrServerError is returned for server-side errors (HTTP 5xx).
	ErrServerError = errors.New("server error: the server encountered an error")
)

// APIError represents an error returned by the Checkvist API.
type APIError struct {
	// StatusCode is the HTTP status code returned by the API.
	StatusCode int
	// Message is a human-readable error message.
	Message string
	// RequestID is the unique identifier for the request, if available.
	RequestID string
	// Err is the underlying sentinel error, if applicable.
	Err error
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("checkvist API error (status %d, request %s): %s", e.StatusCode, e.RequestID, e.Message)
	}
	return fmt.Sprintf("checkvist API error (status %d): %s", e.StatusCode, e.Message)
}

// Unwrap returns the underlying error for use with errors.Is() and errors.As().
func (e *APIError) Unwrap() error {
	return e.Err
}

// NewAPIError creates an APIError from an HTTP response.
// It automatically maps the status code to the appropriate sentinel error.
func NewAPIError(resp *http.Response, message string) *APIError {
	if message == "" {
		message = http.StatusText(resp.StatusCode)
	}

	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		Message:    message,
		RequestID:  resp.Header.Get("X-Request-Id"),
	}

	// Map status codes to sentinel errors
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		apiErr.Err = ErrUnauthorized
	case http.StatusNotFound:
		apiErr.Err = ErrNotFound
	case http.StatusTooManyRequests:
		apiErr.Err = ErrRateLimited
	case http.StatusBadRequest:
		apiErr.Err = ErrBadRequest
	default:
		if resp.StatusCode >= 500 {
			apiErr.Err = ErrServerError
		}
	}

	return apiErr
}
