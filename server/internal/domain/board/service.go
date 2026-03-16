package board

import (
	"errors"
	"server/internal/pkg/response"
	"server/internal/server/middleware"
	"time"

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

// CalculateNextResetAt calculates the next reset time based on the schedule type,
// advancing from the given base time. Pass time.Now().UTC() for initial creation;
// pass the board's current NextResetAt when advancing after a reset so the cadence
// stays anchored to the original schedule rather than drifting with wall clock.
func CalculateNextResetAt(schedule *BoardSchedule, from time.Time) *time.Time {
	if schedule == nil {
		return nil
	}

	switch schedule.Type {
	case "interval":
		if schedule.IntervalSeconds != nil {
			nextReset := from.Add(time.Duration(*schedule.IntervalSeconds) * time.Second)
			return &nextReset
		}
	}

	return nil
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
		// Calculate and set the next reset time
		board.NextResetAt = CalculateNextResetAt(board.Schedule, time.Now().UTC())
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

func (s *Service) List(limit *int, cursorStr string) (*response.PaginatedResponse, error) {
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
	var limitValue int
	if limit != nil {
		limitValue = *limit
		if hasNext && len(boards) > 0 {
			last := boards[len(boards)-1]
			c := EncodeCursor(last.CreatedAt, last.ID)
			nextCursor = &c
		}
	}

	return &response.PaginatedResponse{
		Data:       items,
		TotalCount: totalCount,
		Limit:      limitValue,
		HasNext:    hasNext,
		Cursor:     nextCursor,
	}, nil
}
