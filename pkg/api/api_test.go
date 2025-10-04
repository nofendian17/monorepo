package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	api := New()
	require.NotNil(t, api, "New() should not return nil")
}

func TestApi_Success(t *testing.T) {
	api := New()
	w := httptest.NewRecorder()
	ctx := context.Background()
	data := map[string]string{"key": "value"}

	api.Success(ctx, w, data)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status OK")
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Expected Content-Type application/json")

	var response Response
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err, "Failed to decode response")

	assert.Equal(t, StatusSuccess, response.Status, "Expected status success")
	assert.NotNil(t, response.Data, "Expected data in response")
}

func TestApi_Error(t *testing.T) {
	api := New()
	w := httptest.NewRecorder()
	ctx := context.Background()
	apiErr := &Error{
		Code:    "TEST_ERROR",
		Message: "Test error message",
	}

	api.Error(ctx, w, http.StatusBadRequest, apiErr)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status BadRequest")

	var response Response
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err, "Failed to decode response")

	assert.Equal(t, StatusError, response.Status, "Expected status error")
	assert.NotNil(t, response.Error, "Expected error in response")
	assert.Equal(t, "TEST_ERROR", response.Error.Code, "Expected error code TEST_ERROR")
}

func TestApi_BadRequest(t *testing.T) {
	api := New()
	w := httptest.NewRecorder()
	ctx := context.Background()

	api.BadRequest(ctx, w, "Bad request message")

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status BadRequest")

	var response Response
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err, "Failed to decode response")

	assert.Equal(t, "BAD_REQUEST", response.Error.Code, "Expected error code BAD_REQUEST")
}

func TestApi_Unauthorized(t *testing.T) {
	api := New()
	w := httptest.NewRecorder()
	ctx := context.Background()

	api.Unauthorized(ctx, w, "Unauthorized message")

	assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected status Unauthorized")

	var response Response
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err, "Failed to decode response")

	assert.Equal(t, "UNAUTHORIZED", response.Error.Code, "Expected error code UNAUTHORIZED")
}

func TestApi_Forbidden(t *testing.T) {
	api := New()
	w := httptest.NewRecorder()
	ctx := context.Background()

	api.Forbidden(ctx, w, "Forbidden message")

	assert.Equal(t, http.StatusForbidden, w.Code, "Expected status Forbidden")

	var response Response
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err, "Failed to decode response")

	assert.Equal(t, "FORBIDDEN", response.Error.Code, "Expected error code FORBIDDEN")
}

func TestApi_NotFound(t *testing.T) {
	api := New()
	w := httptest.NewRecorder()
	ctx := context.Background()

	api.NotFound(ctx, w, "Not found message")

	assert.Equal(t, http.StatusNotFound, w.Code, "Expected status NotFound")

	var response Response
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err, "Failed to decode response")

	assert.Equal(t, "NOT_FOUND", response.Error.Code, "Expected error code NOT_FOUND")
}

func TestApi_InternalServerError(t *testing.T) {
	api := New()
	w := httptest.NewRecorder()
	ctx := context.Background()

	api.InternalServerError(ctx, w, "Internal server error message")

	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected status InternalServerError")

	var response Response
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err, "Failed to decode response")

	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Error.Code, "Expected error code INTERNAL_SERVER_ERROR")
}

func TestApi_ValidationError(t *testing.T) {
	api := New()
	w := httptest.NewRecorder()
	ctx := context.Background()
	details := []ErrorDetail{
		{Field: "name", Message: "Name is required"},
		{Field: "email", Message: "Email is invalid"},
	}

	api.ValidationError(ctx, w, details)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Expected status UnprocessableEntity")

	var response Response
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err, "Failed to decode response")

	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code, "Expected error code VALIDATION_ERROR")
	assert.Len(t, response.Error.Details, 2, "Expected 2 error details")
}

func TestApi_SuccessWithMeta(t *testing.T) {
	api := New()
	w := httptest.NewRecorder()
	ctx := context.Background()
	data := "test data"
	meta := &Meta{
		Pagination: &Pagination{
			Page:        1,
			Limit:       10,
			Total:       100,
			TotalPages:  10,
			HasNextPage: true,
			HasPrevPage: false,
		},
	}

	api.SuccessWithMeta(ctx, w, data, meta)

	var response Response
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err, "Failed to decode response")

	assert.NotNil(t, response.Meta, "Expected meta in response")
	assert.Equal(t, 1, response.Meta.Pagination.Page, "Expected page 1")
}

func TestApi_getRequestID(t *testing.T) {
	api := &api{}
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, "test-request-id")

	requestID := api.getRequestID(ctx)
	assert.Equal(t, "test-request-id", requestID, "Expected request ID 'test-request-id'")
}

func TestApi_buildResponse(t *testing.T) {
	api := &api{}
	ctx := context.Background()
	data := "test data"
	meta := &Meta{}
	apiErr := &Error{Code: "TEST"}

	response := api.buildResponse(ctx, StatusSuccess, data, meta, apiErr)

	assert.Equal(t, StatusSuccess, response.Status, "Expected status success")
	assert.Equal(t, data, response.Data, "Expected correct data")
	assert.Equal(t, meta, response.Meta, "Expected correct meta")
	assert.Equal(t, apiErr, response.Error, "Expected correct error")
}

func TestApi_SuccessWithCode(t *testing.T) {
	api := New()
	w := httptest.NewRecorder()
	ctx := context.Background()
	data := map[string]string{"key": "value"}

	api.SuccessWithCode(ctx, w, data)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status OK")
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Expected Content-Type application/json")

	var response Response
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err, "Failed to decode response")

	assert.Equal(t, StatusSuccess, response.Status, "Expected status success")
	assert.NotNil(t, response.Data, "Expected data in response")
}

func TestApi_SuccessWithCodeAndMeta(t *testing.T) {
	api := New()
	w := httptest.NewRecorder()
	ctx := context.Background()
	data := "test data"
	meta := &Meta{
		Pagination: &Pagination{
			Page:        1,
			Limit:       10,
			Total:       100,
			TotalPages:  10,
			HasNextPage: true,
			HasPrevPage: false,
		},
	}

	api.SuccessWithCodeAndMeta(ctx, w, data, meta)

	var response Response
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err, "Failed to decode response")

	assert.NotNil(t, response.Meta, "Expected meta in response")
	assert.Equal(t, 1, response.Meta.Pagination.Page, "Expected page 1")
	assert.Equal(t, StatusSuccess, response.Status, "Expected status success")
}
