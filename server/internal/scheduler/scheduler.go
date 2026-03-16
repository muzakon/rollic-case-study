package scheduler

import (
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"server/internal/domain/board"
	"server/internal/domain/score"
)

// Scheduler periodically checks for boards whose reset window has elapsed,
// clears their scores, and advances their next_reset_at.
type Scheduler struct {
	s         gocron.Scheduler
	boardRepo *board.Repository
	scoreRepo *score.Repository
	log       *zerolog.Logger
}

// New creates a Scheduler and registers the board-reset job.
// The job runs every minute using singleton mode (LimitModeReschedule) so a
// slow execution is skipped rather than stacked.
func New(boardRepo *board.Repository, scoreRepo *score.Repository, log *zerolog.Logger) (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	sch := &Scheduler{
		s:         s,
		boardRepo: boardRepo,
		scoreRepo: scoreRepo,
		log:       log,
	}

	_, err = s.NewJob(
		gocron.CronJob("* * * * *", false),
		gocron.NewTask(sch.resetDueBoards),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
		gocron.WithStartAt(gocron.WithStartImmediately()),
	)
	if err != nil {
		_ = s.Shutdown()
		return nil, err
	}

	return sch, nil
}

// Start begins the scheduler. It is non-blocking.
func (sch *Scheduler) Start() {
	sch.s.Start()
	sch.log.Info().Msg("board reset scheduler started")
}

// Shutdown gracefully stops the scheduler. Call this during application shutdown.
func (sch *Scheduler) Shutdown() error {
	return sch.s.Shutdown()
}

// resetDueBoards is the job function. It fetches all boards whose next_reset_at
// has elapsed, deletes their scores in one query, and updates their next_reset_at
// in a single transaction.
func (sch *Scheduler) resetDueBoards() {
	now := time.Now().UTC()
	sch.log.Debug().Time("at", now).Msg("scheduler: checking for due boards")

	boards, err := sch.boardRepo.FindDueBoards(now)
	if err != nil {
		sch.log.Error().Err(err).Msg("scheduler: failed to query due boards")
		return
	}

	if len(boards) == 0 {
		sch.log.Debug().Msg("scheduler: no boards due for reset")
		return
	}

	sch.log.Info().Int("count", len(boards)).Msg("scheduler: resetting due boards")

	boardIDs := make([]uuid.UUID, len(boards))
	entries := make([]board.ResetEntry, len(boards))
	for i := range boards {
		boardIDs[i] = boards[i].ID
		entries[i] = board.ResetEntry{
			ID:          boards[i].ID,
			NextResetAt: board.CalculateNextResetAt(boards[i].Schedule, *boards[i].NextResetAt),
		}
	}

	if err := sch.scoreRepo.DeleteByBoardIDs(boardIDs); err != nil {
		sch.log.Error().Err(err).Msg("scheduler: failed to delete scores")
		return
	}

	if err := sch.boardRepo.BatchUpdateNextResetAt(entries); err != nil {
		sch.log.Error().Err(err).Msg("scheduler: failed to update next_reset_at")
		return
	}
}
