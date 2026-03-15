package score

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Service struct {
	repo *Repository
	log  *zerolog.Logger
}

func NewService(repo *Repository, log *zerolog.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// GetTopScores returns the top n scores for a board as response DTOs.
func (s *Service) GetTopScores(boardID uuid.UUID, n int) ([]ScoreResponse, error) {
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
