package middleware

import (
	"encoding/json"
	"errors"

	appvalidator "server/internal/pkg/validator"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

// AppError is a custom error type that handlers can return to control HTTP status and message.
type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"error"`
}

func (e *AppError) Error() string {
	return e.Message
}

// NewAppError creates a new AppError.
func NewAppError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// ErrNotFound is a convenience for 404 errors.
func ErrNotFound(resource string) *AppError {
	return NewAppError(fiber.StatusNotFound, resource+" not found")
}

// ErrBadRequest is a convenience for 400 errors.
func ErrBadRequest(message string) *AppError {
	return NewAppError(fiber.StatusBadRequest, message)
}

// GlobalErrorHandler is the Fiber error handler that catches all returned errors
// and formats them into consistent JSON responses.
func GlobalErrorHandler(c fiber.Ctx, err error) error {
	// JSON unmarshaling errors (type mismatches, syntax errors)
	var jsonErr *json.UnmarshalTypeError
	var syntaxErr *json.SyntaxError
	if errors.As(err, &jsonErr) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format",
			"details": []appvalidator.ValidationError{{
				Field:   jsonErr.Field,
				Message: "invalid type or format",
			}},
		})
	}
	if errors.As(err, &syntaxErr) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON syntax",
			"details": []appvalidator.ValidationError{{
				Field:   "unknown",
				Message: "malformed JSON",
			}},
		})
	}

	// Validation errors from go-playground/validator
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": appvalidator.FormatErrors(ve),
		})
	}

	// Custom application errors
	var appErr *AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.Code).JSON(fiber.Map{
			"error": appErr.Message,
		})
	}

	// Fiber's own errors (e.g., 404 from unmatched routes, body parse errors)
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		msg := fiberErr.Message
		if fiberErr.Code == fiber.StatusUnprocessableEntity {
			msg = "Request body is empty or contains invalid JSON. Please send a valid JSON object."
		}
		return c.Status(fiberErr.Code).JSON(fiber.Map{
			"error": msg,
		})
	}

	// Fallback: unexpected internal server error
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": "Internal server error",
	})
}
