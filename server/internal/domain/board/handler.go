package board

import (
	"server/internal/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Handler struct {
	service *Service
	log     *zerolog.Logger
}

func NewHandler(db *gorm.DB, log *zerolog.Logger) *Handler {
	repo := NewRepository(db)
	svc := NewService(repo, log)
	return &Handler{service: svc, log: log}
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
	req := new(CreateBoardRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	board, err := h.service.Create(req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create board")
	}

	return response.Created(c, ToBoardResponse(board))
}
