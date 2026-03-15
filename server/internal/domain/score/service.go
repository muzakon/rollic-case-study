package score

import (
	"server/internal/server/middleware"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type BoardChecker interface {
	Exists(id uuid.UUID) (bool, error)
}

type Service struct {
	repo         *Repository
	boardChecker BoardChecker
	log          *zerolog.Logger
}

func NewService(repo *Repository, boardChecker BoardChecker, log *zerolog.Logger) *Service {
	return &Service{repo: repo, boardChecker: boardChecker, log: log}
}

// GetTopScores returns the top n scores for a board as response DTOs.
func (s *Service) GetTopScores(boardID uuid.UUID, n int) ([]ScoreResponse, error) {
	exists, err := s.boardChecker.Exists(boardID)
	if err != nil {
		s.log.Error().Err(err).Str("boardId", boardID.String()).Msg("failed to check board existence")
		return nil, err
	}
	if !exists {
		return nil, middleware.ErrNotFound("Board")
	}

	scores, err := s.repo.GetTopScores(boardID, n)
	if err != nil {
		s.log.Error().Err(err).Str("boardId", boardID.String()).Msg("failed to get top scores")
		return nil, err
	}

	result := make([]ScoreResponse, len(scores))
	for i, sc := range scores {
		result[i] = ScoreResponse{
			UserID: sc.UserID,
			Score:  sc.ScoreValue,
		}
	}
	return result, nil
}
