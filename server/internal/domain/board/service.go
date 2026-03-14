package board

import (
	"errors"
	"server/internal/pkg/response"
	"server/internal/server/middleware"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
	log  *zerolog.Logger
}

func NewService(repo *Repository, log *zerolog.Logger) *Service {
	return &Service{repo: repo, log: log}
}

func (s *Service) Create(req *CreateBoardRequest) (*Board, error) {
	board := &Board{
		Name:        req.Name,
		Description: req.Description,
	}

	if req.Schedule != nil {
		board.Schedule = &BoardSchedule{
			Type:            req.Schedule.Type,
			IntervalSeconds: req.Schedule.IntervalSeconds,
		}
	}

	if err := s.repo.Create(board); err != nil {
		s.log.Error().Err(err).Msg("failed to create board")
		return nil, err
	}

	return board, nil
}

func (s *Service) GetByID(id uuid.UUID) (*Board, error) {
	board, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, middleware.ErrNotFound("Board")
		}
		s.log.Error().Err(err).Str("boardId", id.String()).Msg("failed to get board")
		return nil, err
	}
	return board, nil
}

func (s *Service) List(limit int, cursorStr string) (*response.PaginatedResponse, error) {
	totalCount, err := s.repo.Count()
	if err != nil {
		s.log.Error().Err(err).Msg("failed to count boards")
		return nil, err
	}

	var cursor *Cursor
	if cursorStr != "" {
		cursor, err = DecodeCursor(cursorStr)
		if err != nil {
			return nil, err
		}
	}

	boards, hasNext, err := s.repo.List(limit, cursor)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to list boards")
		return nil, err
	}

	items := make([]BoardListItem, len(boards))
	for i := range boards {
		items[i] = ToBoardListItem(&boards[i])
	}

	var nextCursor *string
	if hasNext && len(boards) > 0 {
		last := boards[len(boards)-1]
		c := EncodeCursor(last.CreatedAt, last.ID)
		nextCursor = &c
	}

	return &response.PaginatedResponse{
		Data:       items,
		TotalCount: totalCount,
		Limit:      limit,
		HasNext:    hasNext,
		Cursor:     nextCursor,
	}, nil
}
