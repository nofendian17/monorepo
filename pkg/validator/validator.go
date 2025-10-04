package validator

import (
	"github.com/go-playground/validator/v10"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Validator defines the interface for validation operations
type Validator interface {
	ValidateStruct(s any) map[string]string
}

// validatorImpl implements the Validator interface
type validatorImpl struct {
	validate *validator.Validate
}

// NewValidator creates a new instance of the go-playground validator
func NewValidator() Validator {
	return &validatorImpl{
		validate: validator.New(),
	}
}

// ValidateStruct validates a struct and returns field-specific errors
func (v *validatorImpl) ValidateStruct(s any) map[string]string {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	validationErrors := make(map[string]string)
	for _, fieldErr := range err.(validator.ValidationErrors) {
		fieldName := prettifyFieldName(fieldErr.Field())
		validationErrors[fieldErr.Field()] = formatValidationError(fieldErr, fieldName)
	}

	return validationErrors
}

// ValidateStruct validates a struct and returns field-specific errors (package-level function for backward compatibility)
func ValidateStruct(s any) map[string]string {
	v := NewValidator()
	return v.ValidateStruct(s)
}

// formatValidationError returns a more descriptive error message based on the validation tag
func formatValidationError(err validator.FieldError, fieldName string) string {
	switch err.Tag() {
	case "required":
		return fieldName + " is required"
	case "email":
		return fieldName + " must be a valid email address"
	case "min":
		return fieldName + " must be at least " + err.Param() + " characters long"
	case "max":
		return fieldName + " must be at most " + err.Param() + " characters long"
	case "len":
		return fieldName + " must be exactly " + err.Param() + " characters long"
	case "numeric":
		return fieldName + " must be a numeric value"
	case "alpha":
		return fieldName + " must contain only letters"
	case "alphanum":
		return fieldName + " must contain only letters and numbers"
	case "eq":
		return fieldName + " must be equal to " + err.Param()
	case "ne":
		return fieldName + " must not be equal to " + err.Param()
	case "lt":
		return fieldName + " must be less than " + err.Param()
	case "lte":
		return fieldName + " must be less than or equal to " + err.Param()
	case "gt":
		return fieldName + " must be greater than " + err.Param()
	case "gte":
		return fieldName + " must be greater than or equal to " + err.Param()
	case "oneof":
		return fieldName + " must be one of the following: " + err.Param()
	default:
		return fieldName + " is invalid"
	}
}

// prettifyFieldName turns a camelCase or PascalCase field into a human-readable string
func prettifyFieldName(field string) string {
	var result []rune
	for i, r := range field {
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Check if the previous character is lowercase (indicating a camelCase transition)
			if i > 0 && field[i-1] >= 'a' && field[i-1] <= 'z' {
				result = append(result, ' ')
			}
		}
		result = append(result, r)
	}
	return cases.Title(language.Und, cases.NoLower).String(string(result))
}
