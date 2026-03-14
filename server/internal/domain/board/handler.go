package board

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

// List returns all boards.
// GET /api/v1/boards
func (h *Handler) List(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "list boards"})
}

// Get returns a single board by ID.
// GET /api/v1/boards/:boardId
func (h *Handler) Get(c fiber.Ctx) error {
	_ = c.Params("boardId")
	return c.JSON(fiber.Map{"message": "get board"})
}

// Create creates a new board.
// POST /api/v1/boards
func (h *Handler) Create(c fiber.Ctx) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "create board"})
}
