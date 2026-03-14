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
	field := jsonFieldName(fe)

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("'%s' is required", field)
	case "min":
		return fmt.Sprintf("'%s' must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("'%s' must be at most %s characters", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("'%s' must be one of: %s", field, fe.Param())
	case "gt":
		return fmt.Sprintf("'%s' must be greater than %s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("'%s' must be greater than or equal to %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("'%s' must be less than or equal to %s", field, fe.Param())
	case "uuid":
		return fmt.Sprintf("'%s' must be a valid UUID", field)
	case "required_if":
		return fmt.Sprintf("'%s' is required for this configuration", field)
	default:
		return fmt.Sprintf("'%s' failed validation: %s", field, fe.Tag())
	}
}
