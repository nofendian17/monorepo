package httpclient

import (
	"log/slog"
	"net/http"
	"time"
)

// Option is a function that configures a Client
type Option func(*Client)

// WithBaseURL sets the base URL for the client
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithTimeout sets the timeout for requests
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithHeaders sets default headers for all requests
func WithHeaders(headers map[string]string) Option {
	return func(c *Client) {
		if c.headers == nil {
			c.headers = make(map[string]string)
		}
		// Make a copy of the headers to ensure immutability after creation
		for k, v := range headers {
			c.headers[k] = v
		}
	}
}

// WithRetryCount sets the number of retries for failed requests
func WithRetryCount(retryCount int) Option {
	return func(c *Client) {
		c.retryCount = retryCount
	}
}

// WithHTTPClient allows using a custom http.Client
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

// WithLogger adds a slog logger to the client for request/response logging
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}
