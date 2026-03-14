package score

import (
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Handler struct {
	db  *gorm.DB
	log *zerolog.Logger
}

func NewHandler(db *gorm.DB, log *zerolog.Logger) *Handler {
	return &Handler{db: db, log: log}
}

// List returns the top N scores for a board.
// GET /api/v1/boards/:boardId/scores?n=10
func (h *Handler) List(c fiber.Ctx) error {
	_ = c.Params("boardId")
	_ = c.Query("n", "10")
	return c.JSON(fiber.Map{"message": "list scores"})
}

// Submit submits a score for a board.
// POST /api/v1/boards/:boardId/scores
func (h *Handler) Submit(c fiber.Ctx) error {
	_ = c.Params("boardId")
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "submit score"})
}

// Surroundings returns scores around a specific user.
// GET /api/v1/boards/:boardId/scores/:userId/surroundings?n=5
func (h *Handler) Surroundings(c fiber.Ctx) error {
	_ = c.Params("boardId")
	_ = c.Params("userId")
	_ = c.Query("n", "5")
	return c.JSON(fiber.Map{"message": "get surroundings"})
}
