package score

import (
	"strconv"

	"server/internal/domain/board"
	"server/internal/domain/user"
	"server/internal/pkg/response"
	"server/internal/server/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Handler struct {
	service *Service
	log     *zerolog.Logger
}

func NewHandler(db *gorm.DB, log *zerolog.Logger) *Handler {
	repo := NewRepository(db)
	boardRepo := board.NewRepository(db)
	userRepo := user.NewRepository(db)
	service := NewService(repo, boardRepo, userRepo, log)
	return &Handler{service: service, log: log}
}

// List returns the top N scores for a board.
// GET /api/v1/boards/:boardId/scores?n=10
func (h *Handler) List(c fiber.Ctx) error {
	boardID, err := uuid.Parse(c.Params("boardId"))
	if err != nil {
		return middleware.ErrBadRequest("invalid board id")
	}

	nStr := c.Query("n", "10")
	n, err := strconv.Atoi(nStr)
	if err != nil || n < 1 {
		return middleware.ErrBadRequest("Invalid value for n")
	}

	scores, err := h.service.GetTopScores(boardID, n)
	if err != nil {
		return err
	}

	return response.OK(c, scores)
}

// Submit submits a score for a board.
// POST /api/v1/boards/:boardId/scores
func (h *Handler) Submit(c fiber.Ctx) error {
	_ = c.Params("boardId")
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "submit score"})
}

// Seed populates a board with n mock users and random scores.
// POST /api/v1/boards/:boardId/scores/seed
func (h *Handler) Seed(c fiber.Ctx) error {
	boardID, err := uuid.Parse(c.Params("boardId"))
	if err != nil {
		return middleware.ErrBadRequest("invalid board id")
	}

	var req SeedRequest
	if err := c.Bind().JSON(&req); err != nil {
		return err
	}

	result, err := h.service.Seed(boardID, req.N)
	if err != nil {
		return err
	}

	return response.Created(c, result)
}

// Surroundings returns scores around a specific user.
// GET /api/v1/boards/:boardId/scores/:userId/surroundings?n=5
func (h *Handler) Surroundings(c fiber.Ctx) error {
	boardID, err := uuid.Parse(c.Params("boardId"))
	if err != nil {
		return middleware.ErrBadRequest("invalid board id")
	}

	userID := c.Params("userId")
	if userID == "" {
		return middleware.ErrBadRequest("user id is required")
	}

	nStr := c.Query("n", "5")
	n, err := strconv.Atoi(nStr)
	if err != nil || n < 1 {
		return middleware.ErrBadRequest("Invalid value for n")
	}

	result, err := h.service.GetSurroundings(boardID, userID, n)
	if err != nil {
		return err
	}

	return response.OK(c, result)
}
