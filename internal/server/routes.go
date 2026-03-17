package server

import (
	"server/internal/config"
	"server/internal/domain/board"
	"server/internal/domain/score"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// RegisterRoutes wires all API v1 route groups to their respective handlers.
func RegisterRoutes(app *fiber.App, log *zerolog.Logger, db *gorm.DB, cfg *config.Config) {
	boardHandler := board.NewHandler(db, log, cfg)
	scoreHandler := score.NewHandler(db, log)

	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Board routes
	boards := v1.Group("/boards")
	boards.Get("/", boardHandler.List)
	boards.Post("/", boardHandler.Create)
	boards.Get("/:boardId", boardHandler.Get)

	// Score routes (nested under boards)
	scores := boards.Group("/:boardId/scores")
	scores.Get("/", scoreHandler.List)
	scores.Post("/", scoreHandler.Submit)
	scores.Post("/seed", scoreHandler.Seed)
	scores.Get("/:userId/surroundings", scoreHandler.Surroundings)
}
