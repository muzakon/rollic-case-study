package server

import (
	"server/internal/config"
	appvalidator "server/internal/pkg/validator"
	"server/internal/server/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// New creates a configured Fiber application with middleware and routes registered.
func New(cfg *config.Config, log *zerolog.Logger, db *gorm.DB) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:         cfg.App.Name,
		BodyLimit:       1 * 1024 * 1024, // 1MB
		StructValidator: appvalidator.New(),
		ErrorHandler:    middleware.GlobalErrorHandler,
	})

	app.Use(middleware.RequestLogger(log))

	RegisterRoutes(app, log, db, cfg)

	return app
}
