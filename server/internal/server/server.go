package server

import (
	"server/internal/config"
	appvalidator "server/internal/pkg/validator"
	"server/internal/server/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func New(cfg *config.Config, log *zerolog.Logger, db *gorm.DB) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:         cfg.App.Name,
		BodyLimit:       1 * 1024 * 1024, // 1MB
		StructValidator: appvalidator.New(),
		ErrorHandler:    middleware.GlobalErrorHandler,
	})

	RegisterRoutes(app, log, db, cfg)

	return app
}
