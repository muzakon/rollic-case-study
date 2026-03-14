package board

import "github.com/google/uuid"

// --- Request DTOs ---

type CreateBoardRequest struct {
	Name        string                 `json:"name" validate:"required,min=2,max=255"`
	Description string                 `json:"description" validate:"max=1000"`
	Schedule    *CreateScheduleRequest `json:"schedule" validate:"omitempty"`
}

type CreateScheduleRequest struct {
	Type            string `json:"type" validate:"required,oneof=interval daily weekly monthly"`
	IntervalSeconds *int   `json:"intervalSeconds" validate:"required_if=Type interval,omitempty,gt=0"`
}

// --- Response DTOs ---

type BoardResponse struct {
	ID          uuid.UUID          `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Schedule    *ScheduleResponse  `json:"schedule"`
}

type ScheduleResponse struct {
	Type            string `json:"type"`
	IntervalSeconds *int   `json:"intervalSeconds,omitempty"`
}

// ToBoardResponse converts a Board model to its API response.
func ToBoardResponse(b *Board) *BoardResponse {
	resp := &BoardResponse{
		ID:          b.ID,
		Name:        b.Name,
		Description: b.Description,
	}

	if b.Schedule != nil {
		resp.Schedule = &ScheduleResponse{
			Type:            b.Schedule.Type,
			IntervalSeconds: b.Schedule.IntervalSeconds,
		}
	}

	return resp
}
