package board

import "github.com/rs/zerolog"

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
