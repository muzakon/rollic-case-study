package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type StructValidator struct {
	validate *validator.Validate
}

func New() *StructValidator {
	return &StructValidator{validate: validator.New(validator.WithRequiredStructEnabled())}
}

// Validate implements fiber.StructValidator interface.
func (v *StructValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

// ValidationError represents a single field validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FormatErrors takes a validator.ValidationErrors and returns user-friendly messages.
func FormatErrors(err error) []ValidationError {
	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		return []ValidationError{{Field: "unknown", Message: err.Error()}}
	}

	errors := make([]ValidationError, 0, len(ve))
	for _, fe := range ve {
		errors = append(errors, ValidationError{
			Field:   jsonFieldName(fe),
			Message: msgForTag(fe),
		})
	}
	return errors
}

func jsonFieldName(fe validator.FieldError) string {
	return strings.ToLower(fe.Field())
}

func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "field is required"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", fe.Param())
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", fe.Param())
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", fe.Param())
	case "uuid":
		return "must be a valid UUID"
	case "required_if":
		return "field is required for this configuration"
	default:
		return fmt.Sprintf("failed validation: %s", fe.Tag())
	}
}
