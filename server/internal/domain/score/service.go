package score

import (
	"errors"
	"slices"

	"server/internal/server/middleware"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type BoardChecker interface {
	Exists(id uuid.UUID) (bool, error)
}

type UserChecker interface {
	Exists(id string) (bool, error)
}

type Service struct {
	repo         *Repository
	boardChecker BoardChecker
	userChecker  UserChecker
	log          *zerolog.Logger
}

func NewService(repo *Repository, boardChecker BoardChecker, userChecker UserChecker, log *zerolog.Logger) *Service {
	return &Service{repo: repo, boardChecker: boardChecker, userChecker: userChecker, log: log}
}

func (s *Service) checkBoardExists(boardID uuid.UUID) error {
	exists, err := s.boardChecker.Exists(boardID)
	if err != nil {
		s.log.Error().Err(err).Str("boardId", boardID.String()).Msg("failed to check board existence")
		return err
	}
	if !exists {
		return middleware.ErrNotFound("Board")
	}
	return nil
}

// GetTopScores returns the top n scores for a board as response DTOs.
func (s *Service) GetTopScores(boardID uuid.UUID, n int) ([]ScoreResponse, error) {
	if err := s.checkBoardExists(boardID); err != nil {
		return nil, err
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

// GetSurroundings returns the user's score along with n scores above and below.
func (s *Service) GetSurroundings(boardID uuid.UUID, userID string, n int) (*SurroundingsResponse, error) {
	if err := s.checkBoardExists(boardID); err != nil {
		return nil, err
	}

	exists, err := s.userChecker.Exists(userID)
	if err != nil {
		s.log.Error().Err(err).Str("userId", userID).Msg("failed to check user existence")
		return nil, err
	}
	if !exists {
		return nil, middleware.ErrNotFound("User")
	}

	userScore, err := s.repo.GetUserScore(boardID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, middleware.ErrNotFound("Score")
		}
		s.log.Error().Err(err).Str("boardId", boardID.String()).Str("userId", userID).Msg("failed to get user score")
		return nil, err
	}

	above, err := s.repo.GetScoresAbove(boardID, userScore.ScoreValue, userScore.AchievedAt, n)
	if err != nil {
		s.log.Error().Err(err).Str("boardId", boardID.String()).Msg("failed to get scores above")
		return nil, err
	}

	below, err := s.repo.GetScoresBelow(boardID, userScore.ScoreValue, userScore.AchievedAt, n)
	if err != nil {
		s.log.Error().Err(err).Str("boardId", boardID.String()).Msg("failed to get scores below")
		return nil, err
	}

	slices.Reverse(above)

	aboveRes := make([]ScoreResponse, len(above))
	for i, sc := range above {
		aboveRes[i] = ScoreResponse{UserID: sc.UserID, Score: sc.ScoreValue}
	}

	belowRes := make([]ScoreResponse, len(below))
	for i, sc := range below {
		belowRes[i] = ScoreResponse{UserID: sc.UserID, Score: sc.ScoreValue}
	}

	return &SurroundingsResponse{
		User:  ScoreResponse{UserID: userScore.UserID, Score: userScore.ScoreValue},
		Above: aboveRes,
		Below: belowRes,
	}, nil
}
