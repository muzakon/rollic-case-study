package server

import (
	"server/internal/domain/board"
	"server/internal/domain/score"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func RegisterRoutes(app *fiber.App, log *zerolog.Logger, db *gorm.DB) {
	boardHandler := board.NewHandler(db, log)
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
	scores.Get("/:userId/surroundings", scoreHandler.Surroundings)
}
