package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	require.NotNil(t, v, "NewValidator() should not return nil")
}

func TestValidateStruct_Valid(t *testing.T) {
	type TestStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
		Age   int    `validate:"gte=18"`
	}

	v := NewValidator()
	ts := TestStruct{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
	}

	errors := v.ValidateStruct(ts)
	assert.Nil(t, errors, "Expected no validation errors")
}

func TestValidateStruct_Invalid(t *testing.T) {
	type TestStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
		Age   int    `validate:"gte=18"`
	}

	v := NewValidator()
	ts := TestStruct{
		Name:  "",
		Email: "invalid-email",
		Age:   15,
	}

	errors := v.ValidateStruct(ts)
	require.NotNil(t, errors, "Expected validation errors")

	assert.Len(t, errors, 3, "Expected 3 validation errors")

	assert.Contains(t, errors, "Name", "Expected error for Name field")
	assert.Contains(t, errors, "Email", "Expected error for Email field")
	assert.Contains(t, errors, "Age", "Expected error for Age field")
}

func TestValidateStruct_ComprehensiveValidation(t *testing.T) {
	type TestStruct struct {
		Name        string `validate:"required,min=2,max=50"`
		Email       string `validate:"required,email"`
		Age         int    `validate:"gte=18,lte=120"`
		Phone       string `validate:"len=10,numeric"`
		Code        string `validate:"alpha,min=3"`
		UserID      string `validate:"alphanum"`
		Status      string `validate:"oneof=active inactive"`
		Score       int    `validate:"gt=0,lt=100"`
		Count       int    `validate:"eq=5"`
		Description string `validate:"ne=empty"`
	}

	v := NewValidator()

	// Test valid struct
	valid := TestStruct{
		Name:        "John Doe",
		Email:       "john@example.com",
		Age:         25,
		Phone:       "1234567890",
		Code:        "ABC",
		UserID:      "user123",
		Status:      "active",
		Score:       85,
		Count:       5,
		Description: "Some description",
	}

	errors := v.ValidateStruct(valid)
	assert.Nil(t, errors, "Expected no validation errors for valid struct")

	// Test invalid struct with various validation failures
	invalid := TestStruct{
		Name:        "",         // required, min=2
		Email:       "invalid",  // email
		Age:         150,        // lte=120
		Phone:       "123",      // len=10, numeric
		Code:        "A1",       // alpha, min=3
		UserID:      "user@123", // alphanum
		Status:      "pending",  // oneof
		Score:       150,        // lt=100
		Count:       3,          // eq=5
		Description: "empty",    // ne=empty
	}

	errors = v.ValidateStruct(invalid)
	require.NotNil(t, errors, "Expected validation errors")

	// Should have multiple errors
	assert.True(t, len(errors) > 5, "Expected multiple validation errors, got %d", len(errors))
}

func TestValidateStruct_PackageLevel(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"required"`
	}

	ts := TestStruct{Name: ""}

	errors := ValidateStruct(ts)
	require.NotNil(t, errors, "Expected validation errors")

	assert.Len(t, errors, 1, "Expected 1 validation error")
}

func TestPrettifyFieldName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"firstName", "First Name"},
		{"userID", "User ID"},
		{"emailAddress", "Email Address"},
		{"name", "Name"},
	}

	for _, test := range tests {
		result := prettifyFieldName(test.input)
		assert.Equal(t, test.expected, result, "prettifyFieldName(%s) should return %s", test.input, test.expected)
	}
}
