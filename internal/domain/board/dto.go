package board

import (
	"time"

	"github.com/google/uuid"
)

// --- Request DTOs ---

type CreateBoardRequest struct {
	Name        string                 `json:"name" validate:"required,min=2,max=255"`
	Description string                 `json:"description" validate:"max=1000"`
	Schedule    *CreateScheduleRequest `json:"schedule" validate:"omitempty"`
}

type CreateScheduleRequest struct {
	Type            string `json:"type" validate:"required,oneof=interval daily weekly monthly"`
	IntervalSeconds *int   `json:"intervalSeconds" validate:"required_if=Type interval,omitempty,gte=60"`
}

// --- Response DTOs ---

type BoardResponse struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	CreatedAt   time.Time         `json:"createdAt"`
	Schedule    *ScheduleResponse `json:"schedule"`
	NextResetAt *time.Time        `json:"nextResetAt"`
}

type ScheduleResponse struct {
	Type            string `json:"type"`
	IntervalSeconds *int   `json:"intervalSeconds,omitempty"`
}

// BoardListItem is the lightweight representation used in list responses.
type BoardListItem struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

// ToBoardResponse converts a Board model to its full API response.
func ToBoardResponse(b *Board) *BoardResponse {
	resp := &BoardResponse{
		ID:          b.ID,
		Name:        b.Name,
		Description: b.Description,
		CreatedAt:   b.CreatedAt,
		NextResetAt: b.NextResetAt,
	}

	if b.Schedule != nil {
		resp.Schedule = &ScheduleResponse{
			Type:            b.Schedule.Type,
			IntervalSeconds: b.Schedule.IntervalSeconds,
		}
	}

	return resp
}

// ToBoardListItem converts a Board model to a list item.
func ToBoardListItem(b *Board) BoardListItem {
	return BoardListItem{
		ID:          b.ID,
		Name:        b.Name,
		Description: b.Description,
	}
}
