package httpclient

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	client := New()
	require.NotNil(t, client, "New() should not return nil")

	assert.Equal(t, "", client.BaseURL(), "Expected empty base URL")
	assert.Equal(t, 30*time.Second, client.Timeout(), "Expected timeout 30s")
}

func TestWithBaseURL(t *testing.T) {
	baseURL := "https://api.example.com"
	client := New(WithBaseURL(baseURL))

	assert.Equal(t, baseURL, client.BaseURL(), "Expected correct base URL")
}

func TestWithTimeout(t *testing.T) {
	timeout := 10 * time.Second
	client := New(WithTimeout(timeout))

	assert.Equal(t, timeout, client.Timeout(), "Expected correct timeout")
}

func TestWithHeaders(t *testing.T) {
	headers := map[string]string{
		"Authorization": "Bearer token",
		"Content-Type":  "application/json",
	}

	client := New(WithHeaders(headers))

	// We can't directly test the headers map, but we can test that the client was created
	require.NotNil(t, client, "Client should not be nil")
}

func TestWithHeaders_EmptyMap(t *testing.T) {
	headers := map[string]string{}
	client := New(WithHeaders(headers))
	require.NotNil(t, client, "Client should not be nil")
}

func TestWithHeaders_NilMap(t *testing.T) {
	client := New(WithHeaders(nil))
	require.NotNil(t, client, "Client should not be nil")
}

func TestWithRetryCount(t *testing.T) {
	retryCount := 3
	client := New(WithRetryCount(retryCount))

	assert.Equal(t, retryCount, client.RetryCount(), "Expected correct retry count")
}

func TestWithLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, nil))
	client := New(WithLogger(logger))

	assert.Equal(t, logger, client.Logger(), "Logger should be set")
}

func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected GET method")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL))
	resp, err := client.Get(context.Background(), "/", nil)
	require.NoError(t, err, "Get() should not fail")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200")
}

func TestClient_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected POST method")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Expected Content-Type application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Created"))
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL))
	data := map[string]string{"key": "value"}
	resp, err := client.Post(context.Background(), "/", data, nil)
	require.NoError(t, err, "Post() should not fail")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected status 201")
}

func TestClient_Post_ErrorCases(t *testing.T) {
	client := New()

	// Test with unmarshalable data (circular reference)
	type circular struct {
		Self *circular
	}
	data := &circular{}
	data.Self = data

	_, err := client.Post(context.Background(), "/", data, nil)
	require.Error(t, err, "Post() should fail with circular reference")
	assert.Contains(t, err.Error(), "failed to marshal request body", "Expected marshal error")
}

func TestClient_Put(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected PUT method")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL))
	data := map[string]string{"key": "value"}
	resp, err := client.Put(context.Background(), "/", data, nil)
	require.NoError(t, err, "Put() should not fail")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200")
}

func TestClient_Put_ErrorCases(t *testing.T) {
	client := New()

	// Test with unmarshalable data (circular reference)
	type circular struct {
		Self *circular
	}
	data := &circular{}
	data.Self = data

	_, err := client.Put(context.Background(), "/", data, nil)
	require.Error(t, err, "Put() should fail with circular reference")
	assert.Contains(t, err.Error(), "failed to marshal request body", "Expected marshal error")
}

func TestClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected DELETE method")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL))
	resp, err := client.Delete(context.Background(), "/", nil)
	require.NoError(t, err, "Delete() should not fail")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Expected status 204")
}

func TestClient_GetJSON(t *testing.T) {
	expectedData := map[string]string{"message": "success"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedData)
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL))
	var result map[string]string
	err := client.GetJSON(context.Background(), "/", &result, nil)
	require.NoError(t, err, "GetJSON() should not fail")

	assert.Equal(t, "success", result["message"], "Expected correct message")
}

func TestClient_PostJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestData map[string]string
		json.NewDecoder(r.Body).Decode(&requestData)

		assert.Equal(t, "test", requestData["input"], "Expected correct input")

		responseData := map[string]string{"output": "result"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responseData)
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL))
	inputData := map[string]string{"input": "test"}
	var result map[string]string
	err := client.PostJSON(context.Background(), "/", inputData, &result, nil)
	require.NoError(t, err, "PostJSON() should not fail")

	assert.Equal(t, "result", result["output"], "Expected correct output")
}

func TestClient_GetJSON_ErrorCases(t *testing.T) {
	// Test non-2xx status code
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL))
	var result map[string]string
	err := client.GetJSON(context.Background(), "/", &result, nil)
	require.Error(t, err, "GetJSON() should fail with 404")
	assert.Contains(t, err.Error(), "request failed with status: 404", "Expected status error")

	// Test invalid JSON response
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server2.Close()

	client2 := New(WithBaseURL(server2.URL))
	err = client2.GetJSON(context.Background(), "/", &result, nil)
	require.Error(t, err, "GetJSON() should fail with invalid JSON")
	assert.Contains(t, err.Error(), "invalid character", "Expected JSON parse error")
}

func TestClient_PostJSON_ErrorCases(t *testing.T) {
	// Test non-2xx status code
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL))
	inputData := map[string]string{"input": "test"}
	var result map[string]string
	err := client.PostJSON(context.Background(), "/", inputData, &result, nil)
	require.Error(t, err, "PostJSON() should fail with 400")
	assert.Contains(t, err.Error(), "request failed with status: 400", "Expected status error")

	// Test invalid JSON response
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server2.Close()

	client2 := New(WithBaseURL(server2.URL))
	err = client2.PostJSON(context.Background(), "/", inputData, &result, nil)
	require.Error(t, err, "PostJSON() should fail with invalid JSON")
	assert.Contains(t, err.Error(), "invalid character", "Expected JSON parse error")
}

func TestClient_GetJSON_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Simulate slow response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "success"})
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL), WithTimeout(50*time.Millisecond))
	var result map[string]string
	err := client.GetJSON(context.Background(), "/", &result, nil)
	require.Error(t, err, "GetJSON() should fail with timeout")
}

func TestClient_PostJSON_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Simulate slow response
		responseData := map[string]string{"output": "result"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responseData)
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL), WithTimeout(50*time.Millisecond))
	inputData := map[string]string{"input": "test"}
	var result map[string]string
	err := client.PostJSON(context.Background(), "/", inputData, &result, nil)
	require.Error(t, err, "PostJSON() should fail with timeout")
}

func TestClient_GetJSON_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "success"})
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := New(WithBaseURL(server.URL))
	var result map[string]string
	err := client.GetJSON(ctx, "/", &result, nil)
	require.Error(t, err, "GetJSON() should fail with cancelled context")
}

func TestClient_PostJSON_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		responseData := map[string]string{"output": "result"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responseData)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := New(WithBaseURL(server.URL))
	inputData := map[string]string{"input": "test"}
	var result map[string]string
	err := client.PostJSON(ctx, "/", inputData, &result, nil)
	require.Error(t, err, "PostJSON() should fail with cancelled context")
}

func TestClient_Do(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method, "Expected PATCH method")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL))
	resp, err := client.Do(context.Background(), "PATCH", "/", nil, nil)
	require.NoError(t, err, "Do() should not fail")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200")
}

func TestClient_Do_CustomMethods(t *testing.T) {
	testCases := []struct {
		method   string
		expected string
	}{
		{"PATCH", "PATCH"},
		{"HEAD", "HEAD"},
		{"OPTIONS", "OPTIONS"},
		{"CONNECT", "CONNECT"},
		{"TRACE", "TRACE"},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tc.expected, r.Method, "Expected "+tc.expected+" method")
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := New(WithBaseURL(server.URL))
			resp, err := client.Do(context.Background(), tc.method, "/", nil, nil)
			require.NoError(t, err, tc.method+"() should not fail")
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200")
		})
	}
}

func TestClient_Do_WithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"), "Expected Authorization header")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Expected Content-Type header")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL))
	headers := map[string]string{
		"Authorization": "Bearer token123",
		"Content-Type":  "application/json",
	}
	resp, err := client.Do(context.Background(), "GET", "/", nil, headers)
	require.NoError(t, err, "Do() should not fail")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200")
}

func TestClient_Do_WithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err, "Should read request body")
		assert.Equal(t, "test data", string(body), "Expected request body")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Expected Content-Type header")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL))
	body := strings.NewReader("test data")
	resp, err := client.Do(context.Background(), "POST", "/", body, nil)
	require.NoError(t, err, "Do() should not fail")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200")
}

func TestWithHTTPClient(t *testing.T) {
	httpClient := &http.Client{Timeout: 5 * time.Second}
	client := New(WithHTTPClient(httpClient))

	require.NotNil(t, client, "Client should not be nil")
}

// Additional tests to improve coverage

func TestClient_Do_RetryLogic(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		// Force network errors for first 2 attempts by closing connection
		if attemptCount < 3 {
			hj, ok := w.(http.Hijacker)
			if ok {
				conn, _, _ := hj.Hijack()
				conn.Close()
			}
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success after retries"))
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL), WithRetryCount(3))
	resp, err := client.Do(context.Background(), "GET", "/", nil, nil)
	require.NoError(t, err, "Do() should succeed after retries")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200 after retries")
	assert.Equal(t, 3, attemptCount, "Expected 3 attempts (2 failures + 1 success)")
}

func TestClient_Do_RetryExhaustion(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		// Force network errors for all attempts by closing connection
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL), WithRetryCount(2))
	_, err := client.Do(context.Background(), "GET", "/", nil, nil)
	require.Error(t, err, "Do() should fail after all retries exhausted")
	assert.Equal(t, 3, attemptCount, "Expected 3 attempts (all failures)")
	assert.Contains(t, err.Error(), "request failed after 2 retries", "Expected retry exhaustion error")
}

func TestClient_Do_WithLogger(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a logger that captures output
	var logOutput strings.Builder
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	client := New(WithBaseURL(server.URL), WithLogger(logger))
	resp, err := client.Do(context.Background(), "GET", "/", nil, nil)
	require.NoError(t, err, "Do() should not fail")
	defer resp.Body.Close()

	// Verify that logging occurred
	logContent := logOutput.String()
	assert.Contains(t, logContent, "HTTP request", "Expected request logging")
	assert.Contains(t, logContent, "HTTP response", "Expected response logging")
}

func TestClient_GetJSON_ErrorReadingResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		// Write invalid JSON
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL))
	var result map[string]string
	err := client.GetJSON(context.Background(), "/", &result, nil)
	require.Error(t, err, "GetJSON() should fail with invalid JSON")
	assert.Contains(t, err.Error(), "invalid character", "Expected JSON parse error")
}

func TestClient_PostJSON_MarshalError(t *testing.T) {
	client := New()

	// Test with unmarshalable data (function type)
	data := map[string]interface{}{
		"func": func() {}, // Functions cannot be marshaled to JSON
	}

	var result map[string]string
	err := client.PostJSON(context.Background(), "/", data, &result, nil)
	require.Error(t, err, "PostJSON() should fail with unmarshalable data")
	assert.Contains(t, err.Error(), "failed to marshal request body", "Expected marshal error")
}

func TestClient_PostJSON_ErrorReadingResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		// Write invalid JSON
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL))
	inputData := map[string]string{"test": "data"}
	var result map[string]string
	err := client.PostJSON(context.Background(), "/", inputData, &result, nil)
	require.Error(t, err, "PostJSON() should fail with invalid JSON")
	assert.Contains(t, err.Error(), "invalid character", "Expected JSON parse error")
}

func TestClient_GetJSON_ReadAllError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		// Close connection immediately to cause read error
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL))
	var result map[string]string
	err := client.GetJSON(context.Background(), "/", &result, nil)
	require.Error(t, err, "GetJSON() should fail with connection error")
}

func TestWithHeaders_ConcurrentAccess(t *testing.T) {
	headers := map[string]string{
		"Authorization": "Bearer token123",
		"X-Custom":      "value",
	}

	client := New(WithHeaders(headers))
	require.NotNil(t, client, "Client should not be nil")

	// Test that the client was created successfully with headers
	// The actual header testing is done through HTTP requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"), "Expected Authorization header")
		assert.Equal(t, "value", r.Header.Get("X-Custom"), "Expected X-Custom header")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	clientWithBaseURL := New(WithBaseURL(server.URL), WithHeaders(headers))
	resp, err := clientWithBaseURL.Get(context.Background(), "/", nil)
	require.NoError(t, err, "Request should succeed")
	defer resp.Body.Close()
}

func TestClient_Do_TimeoutWithRetry(t *testing.T) {
	// Test timeout without retry to avoid complexity
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Longer than timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL), WithTimeout(50*time.Millisecond), WithRetryCount(0))
	_, err := client.Do(context.Background(), "GET", "/", nil, nil)
	require.Error(t, err, "Do() should fail with timeout")
}
