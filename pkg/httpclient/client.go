package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// HTTPClient defines the interface for HTTP client operations
type HTTPClient interface {
	Get(ctx context.Context, path string, headers map[string]string) (*http.Response, error)
	Post(ctx context.Context, path string, data interface{}, headers map[string]string) (*http.Response, error)
	Put(ctx context.Context, path string, data interface{}, headers map[string]string) (*http.Response, error)
	Delete(ctx context.Context, path string, headers map[string]string) (*http.Response, error)
	GetJSON(ctx context.Context, path string, result interface{}, headers map[string]string) error
	PostJSON(ctx context.Context, path string, data interface{}, result interface{}, headers map[string]string) error
	Do(ctx context.Context, method, path string, body io.Reader, headers map[string]string) (*http.Response, error)
	BaseURL() string
	Timeout() time.Duration
	RetryCount() int
	Logger() *slog.Logger
}

// Client represents an HTTP client with configurable settings
type Client struct {
	client     *http.Client
	baseURL    string
	headers    map[string]string
	timeout    time.Duration
	retryCount int
	logger     *slog.Logger
}

// New creates a new HTTP client with the provided options
func New(opts ...Option) HTTPClient {
	client := &Client{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers:    make(map[string]string),
		timeout:    30 * time.Second,
		retryCount: 0,
	}

	for _, opt := range opts {
		opt(client)
	}

	// Update the client's timeout with the configured timeout
	client.client.Timeout = client.timeout

	// Ensure headers map is properly initialized and immutable after this point
	if client.headers == nil {
		client.headers = make(map[string]string)
	}

	return client
}

// Get performs an HTTP GET request
func (c *Client) Get(ctx context.Context, path string, headers map[string]string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, path, nil, headers)
}

// Post performs an HTTP POST request with JSON data
func (c *Client) Post(ctx context.Context, path string, data interface{}, headers map[string]string) (*http.Response, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return c.do(ctx, http.MethodPost, path, bytes.NewBuffer(body), headers)
}

// Put performs an HTTP PUT request with JSON data
func (c *Client) Put(ctx context.Context, path string, data interface{}, headers map[string]string) (*http.Response, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return c.do(ctx, http.MethodPut, path, bytes.NewBuffer(body), headers)
}

// Delete performs an HTTP DELETE request
func (c *Client) Delete(ctx context.Context, path string, headers map[string]string) (*http.Response, error) {
	return c.do(ctx, http.MethodDelete, path, nil, headers)
}

// do performs an HTTP request with the given method, path, and body
func (c *Client) do(ctx context.Context, method, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type if body is provided
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set default headers - safe for concurrent use since headers are immutable after creation
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// Set additional headers for this specific request
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Log the request if logger is configured
	if c.logger != nil {
		c.logger.Info("HTTP request", "method", method, "url", url, "headers", headers)
	}

	// Perform the request with retries if configured
	var resp *http.Response
	var lastErr error

	for i := 0; i <= c.retryCount; i++ {
		resp, lastErr = c.client.Do(req)
		if lastErr == nil {
			break
		}

		// If this was the last attempt, break and return the error
		if i == c.retryCount {
			break
		}

		// Wait before retrying with exponential backoff and jitter
		backoffDuration := time.Duration(1<<uint(i)) * time.Second
		// Add some jitter to prevent thundering herd
		jitter := time.Duration((i+1)*100) * time.Millisecond
		time.Sleep(backoffDuration + jitter)

		// Log retry attempt if logger is configured
		if c.logger != nil {
			c.logger.Info("Retrying HTTP request", "attempt", i+1, "error", lastErr.Error())
		}
	}

	if lastErr != nil {
		errMsg := fmt.Sprintf("request failed after %d retries", c.retryCount)
		if c.logger != nil {
			c.logger.Error(errMsg, "method", method, "url", url, "error", lastErr)
		}
		return nil, fmt.Errorf("%s: %w", errMsg, lastErr)
	}

	// Log the response if logger is configured
	if c.logger != nil {
		c.logger.Info("HTTP response", "method", method, "url", url, "status", resp.Status, "statusCode", resp.StatusCode)
	}

	return resp, nil
}

// GetJSON performs a GET request and unmarshals the response into the provided interface
func (c *Client) GetJSON(ctx context.Context, path string, result interface{}, headers map[string]string) error {
	resp, err := c.Get(ctx, path, headers)
	if err != nil {
		return err
	}
	defer func() {
		// Best practice: always close the response body, ignoring errors
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			if c.logger != nil {
				c.logger.Error("Failed to read response body", "path", path, "error", err)
			}
			return fmt.Errorf("request failed with status: %d, unable to read body: %w", resp.StatusCode, err)
		}

		if c.logger != nil {
			c.logger.Error("HTTP request failed", "path", path, "status", resp.StatusCode, "body", string(body))
		}
		return fmt.Errorf("request failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Failed to read response body", "path", path, "error", err)
		}
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(responseBody, result); err != nil {
		if c.logger != nil {
			c.logger.Error("Failed to unmarshal response", "path", path, "error", err)
		}
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// PostJSON performs a POST request with JSON data and unmarshals the response into the provided interface
func (c *Client) PostJSON(ctx context.Context, path string, data interface{}, result interface{}, headers map[string]string) error {
	body, err := json.Marshal(data)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Failed to marshal request body", "path", path, "error", err)
		}
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := c.do(ctx, http.MethodPost, path, bytes.NewBuffer(body), headers)
	if err != nil {
		return err
	}
	defer func() {
		// Best practice: always close the response body, ignoring errors
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			if c.logger != nil {
				c.logger.Error("Failed to read response body", "path", path, "error", err)
			}
			return fmt.Errorf("request failed with status: %d, unable to read body: %w", resp.StatusCode, err)
		}

		if c.logger != nil {
			c.logger.Error("HTTP request failed", "path", path, "status", resp.StatusCode, "body", string(body))
		}
		return fmt.Errorf("request failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Failed to read response body", "path", path, "error", err)
		}
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(responseBody, result); err != nil {
		if c.logger != nil {
			c.logger.Error("Failed to unmarshal response", "path", path, "error", err)
		}
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// Do performs an HTTP request with the given method, path, and body
// This is a public method that allows for custom HTTP methods beyond the standard ones
func (c *Client) Do(ctx context.Context, method, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	return c.do(ctx, method, path, body, headers)
}

// BaseURL returns the base URL of the client
func (c *Client) BaseURL() string {
	return c.baseURL
}

// Timeout returns the timeout setting of the client
func (c *Client) Timeout() time.Duration {
	return c.timeout
}

// RetryCount returns the retry count setting of the client
func (c *Client) RetryCount() int {
	return c.retryCount
}

// Logger returns the logger of the client
func (c *Client) Logger() *slog.Logger {
	return c.logger
}
