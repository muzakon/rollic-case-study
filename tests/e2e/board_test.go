package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// boardResponse mirrors the API response for a single board.
type boardResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	CreatedAt   string            `json:"createdAt"`
	Schedule    *scheduleResponse `json:"schedule"`
	NextResetAt *string           `json:"nextResetAt"`
}

type scheduleResponse struct {
	Type            string `json:"type"`
	IntervalSeconds *int   `json:"intervalSeconds,omitempty"`
}

// paginatedBoardResponse mirrors the paginated list response.
type paginatedBoardResponse struct {
	Data       []boardListItem `json:"data"`
	TotalCount int             `json:"totalCount"`
	Limit      int             `json:"limit"`
	HasNext    bool            `json:"hasNext"`
	Cursor     *string         `json:"cursor"`
}

type boardListItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// errorResponse mirrors the standard error response shape.
type errorResponse struct {
	Error   string `json:"error"`
	Details []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	} `json:"details"`
}

const boardsPath = "/api/v1/boards"

// --- POST /api/v1/boards ---

func TestCreateBoard_Success(t *testing.T) {
	cleanBoards(t)

	body := map[string]any{
		"name":        "Weekly Challenge",
		"description": "A weekly leaderboard",
	}

	resp := performRequest(t, http.MethodPost, boardsPath, body)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result boardResponse
	parseJSON(t, resp, &result)

	assert.Equal(t, "Weekly Challenge", result.Name)
	assert.Equal(t, "A weekly leaderboard", result.Description)
	assert.NotEmpty(t, result.ID, "response should contain a UUID")
	assert.NotEmpty(t, result.CreatedAt, "createdAt should be set")
	assert.Nil(t, result.Schedule, "schedule should be nil when not provided")
	assert.Nil(t, result.NextResetAt, "nextResetAt should be nil without schedule")
}

func TestCreateBoard_WithSchedule(t *testing.T) {
	cleanBoards(t)

	interval := 3600
	body := map[string]any{
		"name":        "Hourly Board",
		"description": "Resets every hour",
		"schedule": map[string]any{
			"type":            "interval",
			"intervalSeconds": interval,
		},
	}

	resp := performRequest(t, http.MethodPost, boardsPath, body)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result boardResponse
	parseJSON(t, resp, &result)

	assert.Equal(t, "Hourly Board", result.Name)
	require.NotNil(t, result.Schedule, "schedule should be present")
	assert.Equal(t, "interval", result.Schedule.Type)
	require.NotNil(t, result.Schedule.IntervalSeconds)
	assert.Equal(t, interval, *result.Schedule.IntervalSeconds)
	assert.NotNil(t, result.NextResetAt, "nextResetAt should be set for scheduled boards")
}

func TestCreateBoard_WithDailySchedule(t *testing.T) {
	cleanBoards(t)

	body := map[string]any{
		"name": "Daily Board",
		"schedule": map[string]any{
			"type": "daily",
		},
	}

	resp := performRequest(t, http.MethodPost, boardsPath, body)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result boardResponse
	parseJSON(t, resp, &result)

	assert.Equal(t, "Daily Board", result.Name)
	require.NotNil(t, result.Schedule)
	assert.Equal(t, "daily", result.Schedule.Type)
}

func TestCreateBoard_MissingName(t *testing.T) {
	cleanBoards(t)

	body := map[string]any{
		"description": "A board with no name",
	}

	resp := performRequest(t, http.MethodPost, boardsPath, body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result errorResponse
	parseJSON(t, resp, &result)

	assert.Equal(t, "Validation failed", result.Error)
	require.NotEmpty(t, result.Details)
	assert.Equal(t, "name", result.Details[0].Field)
}

func TestCreateBoard_EmptyBody(t *testing.T) {
	cleanBoards(t)

	req, err := http.NewRequest(http.MethodPost, boardsPath, nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testApp.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateBoard_NameTooShort(t *testing.T) {
	cleanBoards(t)

	body := map[string]any{
		"name": "A", // min=2
	}

	resp := performRequest(t, http.MethodPost, boardsPath, body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result errorResponse
	parseJSON(t, resp, &result)

	assert.Equal(t, "Validation failed", result.Error)
	require.NotEmpty(t, result.Details)
	assert.Equal(t, "name", result.Details[0].Field)
	assert.Contains(t, result.Details[0].Message, "at least 2")
}

func TestCreateBoard_InvalidScheduleType(t *testing.T) {
	cleanBoards(t)

	body := map[string]any{
		"name": "Bad Schedule Board",
		"schedule": map[string]any{
			"type": "yearly", // not in oneof: interval daily weekly monthly
		},
	}

	resp := performRequest(t, http.MethodPost, boardsPath, body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result errorResponse
	parseJSON(t, resp, &result)

	assert.Equal(t, "Validation failed", result.Error)
	require.NotEmpty(t, result.Details)
	assert.Equal(t, "type", result.Details[0].Field)
	assert.Contains(t, result.Details[0].Message, "one of")
}

func TestCreateBoard_IntervalScheduleMissingSeconds(t *testing.T) {
	cleanBoards(t)

	body := map[string]any{
		"name": "Missing Interval Board",
		"schedule": map[string]any{
			"type": "interval",
			// intervalSeconds missing — required_if=Type interval
		},
	}

	resp := performRequest(t, http.MethodPost, boardsPath, body)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateBoard_IntervalTooSmall(t *testing.T) {
	cleanBoards(t)

	body := map[string]any{
		"name": "Tiny Interval Board",
		"schedule": map[string]any{
			"type":            "interval",
			"intervalSeconds": 10, // gte=60 required
		},
	}

	resp := performRequest(t, http.MethodPost, boardsPath, body)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateBoard_Persisted(t *testing.T) {
	cleanBoards(t)

	body := map[string]any{
		"name":        "Persistent Board",
		"description": "Should survive a GET after POST",
	}

	createResp := performRequest(t, http.MethodPost, boardsPath, body)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var created boardResponse
	parseJSON(t, createResp, &created)
	require.NotEmpty(t, created.ID)

	getResp := performRequest(t, http.MethodGet, boardsPath+"/"+created.ID, nil)
	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	var fetched boardResponse
	parseJSON(t, getResp, &fetched)

	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, "Persistent Board", fetched.Name)
	assert.Equal(t, "Should survive a GET after POST", fetched.Description)
}

// --- GET /api/v1/boards ---

func TestListBoards_Empty(t *testing.T) {
	cleanBoards(t)

	resp := performRequest(t, http.MethodGet, boardsPath, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result paginatedBoardResponse
	parseJSON(t, resp, &result)

	assert.Equal(t, int64(0), int64(result.TotalCount))
	assert.Empty(t, result.Data)
}

func TestListBoards_Success(t *testing.T) {
	cleanBoards(t)

	// Create two boards
	for _, name := range []string{"Board A", "Board B"} {
		performRequest(t, http.MethodPost, boardsPath, map[string]any{"name": name})
	}

	resp := performRequest(t, http.MethodGet, boardsPath, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result paginatedBoardResponse
	parseJSON(t, resp, &result)

	assert.Equal(t, 2, result.TotalCount)
	assert.Len(t, result.Data, 2)
}

func TestListBoards_Pagination(t *testing.T) {
	cleanBoards(t)

	// Create 3 boards
	for i := range 3 {
		performRequest(t, http.MethodPost, boardsPath, map[string]any{
			"name": fmt.Sprintf("Paginated Board %d", i),
		})
	}

	// Page 1: limit=2
	resp := performRequest(t, http.MethodGet, boardsPath+"?limit=2", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var page1 paginatedBoardResponse
	parseJSON(t, resp, &page1)

	assert.Len(t, page1.Data, 2)
	assert.True(t, page1.HasNext)
	assert.NotNil(t, page1.Cursor)

	// Page 2: use cursor
	resp2 := performRequest(t, http.MethodGet, boardsPath+"?limit=2&cursor="+*page1.Cursor, nil)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var page2 paginatedBoardResponse
	parseJSON(t, resp2, &page2)

	assert.Len(t, page2.Data, 1)
	assert.False(t, page2.HasNext)
}

// --- GET /api/v1/boards/:boardId ---

func TestGetBoard_Success(t *testing.T) {
	cleanBoards(t)

	body := map[string]any{
		"name":        "Get Test Board",
		"description": "To be fetched",
	}
	createResp := performRequest(t, http.MethodPost, boardsPath, body)
	var created boardResponse
	parseJSON(t, createResp, &created)

	resp := performRequest(t, http.MethodGet, boardsPath+"/"+created.ID, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var fetched boardResponse
	parseJSON(t, resp, &fetched)

	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, "Get Test Board", fetched.Name)
	assert.NotEmpty(t, fetched.CreatedAt)
}

func TestGetBoard_NotFound(t *testing.T) {
	cleanBoards(t)

	fakeId := "123e4567-e89b-12d3-a456-426614174000"
	resp := performRequest(t, http.MethodGet, boardsPath+"/"+fakeId, nil)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var result errorResponse
	parseJSON(t, resp, &result)
	assert.Contains(t, result.Error, "not found")
}

func TestGetBoard_InvalidUUID(t *testing.T) {
	cleanBoards(t)

	resp := performRequest(t, http.MethodGet, boardsPath+"/not-a-uuid-string", nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
