package response

import "github.com/gofiber/fiber/v3"

// PaginatedResponse is the standard paginated response envelope.
type PaginatedResponse struct {
	Data       any     `json:"data"`
	TotalCount int64   `json:"totalCount"`
	Limit      int     `json:"limit"`
	HasNext    bool    `json:"hasNext"`
	Cursor     *string `json:"cursor"`
}

// JSON sends a success response with data.
func JSON(c fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(data)
}

// Created sends a 201 response with data.
func Created(c fiber.Ctx, data any) error {
	return JSON(c, fiber.StatusCreated, data)
}

// OK sends a 200 response with data.
func OK(c fiber.Ctx, data any) error {
	return JSON(c, fiber.StatusOK, data)
}
