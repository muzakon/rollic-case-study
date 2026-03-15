package score

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"time"

	"server/internal/server/middleware"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type BoardChecker interface {
	Exists(id uuid.UUID) (bool, error)
}

type Service struct {
	repo      *Repository
	boardRepo BoardChecker
	log       *zerolog.Logger
}

func NewService(repo *Repository, boardRepo BoardChecker, log *zerolog.Logger) *Service {
	return &Service{repo: repo, boardRepo: boardRepo, log: log}
}

func (s *Service) checkBoardExists(boardID uuid.UUID) error {
	exists, err := s.boardRepo.Exists(boardID)
	if err != nil {
		s.log.Error().Err(err).Str("boardId", boardID.String()).Msg("failed to check board existence")
		return err
	}
	if !exists {
		return middleware.ErrNotFound("Board")
	}
	return nil
}

// Submit creates or updates a user's score on a board.
func (s *Service) Submit(boardID uuid.UUID, req *SubmitRequest) (*SubmitResponse, error) {
	if err := s.checkBoardExists(boardID); err != nil {
		return nil, err
	}

	sc := &Score{
		BoardID:    boardID,
		UserID:     req.UserID,
		ScoreValue: req.Score,
		AchievedAt: time.Now(),
	}

	if err := s.repo.Upsert(sc); err != nil {
		s.log.Error().Err(err).Str("boardId", boardID.String()).Str("userId", req.UserID).Msg("failed to upsert score")
		return nil, err
	}

	return &SubmitResponse{
		BoardID: boardID.String(),
		UserID:  sc.UserID,
		Score:   sc.ScoreValue,
	}, nil
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

// Seed creates n mock scores for a board.
func (s *Service) Seed(boardID uuid.UUID, n int) (*SeedResponse, error) {
	if err := s.checkBoardExists(boardID); err != nil {
		return nil, err
	}

	scores := make([]Score, n)
	now := time.Now()

	for i := range n {
		scores[i] = Score{
			BoardID:    boardID,
			UserID:     fmt.Sprintf("user_%d", now.UnixNano()+int64(i)),
			ScoreValue: rand.IntN(10000) + 1,
			AchievedAt: now.Add(-time.Duration(rand.IntN(3600)) * time.Second),
		}
	}

	if err := s.repo.CreateMany(scores); err != nil {
		s.log.Error().Err(err).Str("boardId", boardID.String()).Msg("failed to create mock scores")
		return nil, err
	}

	return &SeedResponse{
		ScoresCreated: n,
	}, nil
}
