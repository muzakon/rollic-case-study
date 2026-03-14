package board

import (
	"server/internal/pkg/response"
	"server/internal/server/middleware"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

const defaultLimit = 20
const maxLimit = 100

type Handler struct {
	service *Service
	log     *zerolog.Logger
}

func NewHandler(db *gorm.DB, log *zerolog.Logger) *Handler {
	repo := NewRepository(db)
	svc := NewService(repo, log)
	return &Handler{service: svc, log: log}
}

// List returns paginated boards.
// GET /api/v1/boards?limit=20&cursor=...
func (h *Handler) List(c fiber.Ctx) error {
	limit := defaultLimit
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= maxLimit {
			limit = parsed
		}
	}
	cursor := c.Query("cursor")

	result, err := h.service.List(limit, cursor)
	if err != nil {
		return middleware.ErrBadRequest(err.Error())
	}

	return response.OK(c, result)
}

// Get returns a single board by ID.
// GET /api/v1/boards/:boardId
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("boardId"))
	if err != nil {
		return middleware.ErrBadRequest("Invalid board ID, must be a valid UUID")
	}

	board, err := h.service.GetByID(id)
	if err != nil {
		return err
	}

	return response.OK(c, ToBoardResponse(board))
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
