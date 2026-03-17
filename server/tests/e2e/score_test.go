package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type scoreResponse struct {
	UserID string `json:"userId"`
	Score  int    `json:"score"`
}

type submitResponse struct {
	BoardID string `json:"boardId"`
	UserID  string `json:"userId"`
	Score   int    `json:"score"`
}

type seedResponse struct {
	ScoresCreated int `json:"scoresCreated"`
}

type surroundingsResponse struct {
	User  scoreResponse   `json:"user"`
	Above []scoreResponse `json:"above"`
	Below []scoreResponse `json:"below"`
}

// createDummyBoard is a helper to create a board for testing scores.
func createDummyBoard(t *testing.T) string {
	t.Helper()
	body := map[string]any{
		"name": "Score Test Board",
	}
	resp := performRequest(t, http.MethodPost, boardsPath, body)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created boardResponse
	parseJSON(t, resp, &created)
	return created.ID
}

// submitScore is a helper to submit a score for a user on a board.
func submitScore(t *testing.T, boardID, userID string, score int) {
	t.Helper()
	body := map[string]any{
		"userId": userID,
		"score":  score,
	}
	path := fmt.Sprintf("%s/%s/scores", boardsPath, boardID)
	resp := performRequest(t, http.MethodPost, path, body)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

// --- POST /api/v1/boards/:boardId/scores ---

func TestSubmitScore_Success(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	body := map[string]any{
		"userId": "player1",
		"score":  1000,
	}
	path := fmt.Sprintf("%s/%s/scores", boardsPath, boardID)
	resp := performRequest(t, http.MethodPost, path, body)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result submitResponse
	parseJSON(t, resp, &result)

	assert.Equal(t, boardID, result.BoardID)
	assert.Equal(t, "player1", result.UserID)
	assert.Equal(t, 1000, result.Score)
}

func TestSubmitScore_OverwritesPreviousScore(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	// Submit initial score
	submitScore(t, boardID, "player1", 500)

	// Overwrite with higher score
	submitScore(t, boardID, "player1", 1500)

	// Verify via top scores — should show 1500, not both
	path := fmt.Sprintf("%s/%s/scores?n=10", boardsPath, boardID)
	resp := performRequest(t, http.MethodGet, path, nil)

	var scores []scoreResponse
	parseJSON(t, resp, &scores)

	require.Len(t, scores, 1, "should have exactly one entry per user")
	assert.Equal(t, 1500, scores[0].Score)
}

func TestSubmitScore_OverwriteWithLowerScore(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	submitScore(t, boardID, "player1", 2000)
	submitScore(t, boardID, "player1", 100)

	path := fmt.Sprintf("%s/%s/scores?n=10", boardsPath, boardID)
	resp := performRequest(t, http.MethodGet, path, nil)

	var scores []scoreResponse
	parseJSON(t, resp, &scores)

	require.Len(t, scores, 1)
	assert.Equal(t, 100, scores[0].Score, "score should be overwritten, not max")
}

func TestSubmitScore_InvalidBoardID(t *testing.T) {
	cleanBoards(t)

	body := map[string]any{
		"userId": "player1",
		"score":  1000,
	}
	path := fmt.Sprintf("%s/%s/scores", boardsPath, "invalid-uuid")
	resp := performRequest(t, http.MethodPost, path, body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSubmitScore_BoardNotFound(t *testing.T) {
	cleanBoards(t)

	fakeID := "123e4567-e89b-12d3-a456-426614174000"
	body := map[string]any{
		"userId": "player1",
		"score":  1000,
	}
	path := fmt.Sprintf("%s/%s/scores", boardsPath, fakeID)
	resp := performRequest(t, http.MethodPost, path, body)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSubmitScore_MissingBody(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	path := fmt.Sprintf("%s/%s/scores", boardsPath, boardID)

	req, err := http.NewRequest(http.MethodPost, path, nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testApp.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSubmitScore_MissingUserId(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	body := map[string]any{
		"score": 1000,
	}
	path := fmt.Sprintf("%s/%s/scores", boardsPath, boardID)
	resp := performRequest(t, http.MethodPost, path, body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSubmitScore_NegativeScore(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	body := map[string]any{
		"userId": "player2",
		"score":  -50,
	}
	path := fmt.Sprintf("%s/%s/scores", boardsPath, boardID)
	resp := performRequest(t, http.MethodPost, path, body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSubmitScore_ZeroScore(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	body := map[string]any{
		"userId": "player_zero",
		"score":  0,
	}
	path := fmt.Sprintf("%s/%s/scores", boardsPath, boardID)
	resp := performRequest(t, http.MethodPost, path, body)

	// gt=0 means score must be > 0, so 0 is invalid
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// --- GET /api/v1/boards/:boardId/scores ---

func TestListScores_Empty(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	path := fmt.Sprintf("%s/%s/scores?n=10", boardsPath, boardID)
	resp := performRequest(t, http.MethodGet, path, nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var scores []scoreResponse
	parseJSON(t, resp, &scores)

	assert.Empty(t, scores)
}

func TestListScores_OrderedDescending(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	submitScore(t, boardID, "low", 100)
	submitScore(t, boardID, "mid", 500)
	submitScore(t, boardID, "high", 1000)

	path := fmt.Sprintf("%s/%s/scores?n=10", boardsPath, boardID)
	resp := performRequest(t, http.MethodGet, path, nil)

	var scores []scoreResponse
	parseJSON(t, resp, &scores)

	require.Len(t, scores, 3)
	assert.Equal(t, "high", scores[0].UserID)
	assert.Equal(t, 1000, scores[0].Score)
	assert.Equal(t, "mid", scores[1].UserID)
	assert.Equal(t, 500, scores[1].Score)
	assert.Equal(t, "low", scores[2].UserID)
	assert.Equal(t, 100, scores[2].Score)
}

func TestListScores_TieBreaking(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	// First player scores 500
	submitScore(t, boardID, "first", 500)
	time.Sleep(10 * time.Millisecond)
	// Second player scores 500 later
	submitScore(t, boardID, "second", 500)

	path := fmt.Sprintf("%s/%s/scores?n=10", boardsPath, boardID)
	resp := performRequest(t, http.MethodGet, path, nil)

	var scores []scoreResponse
	parseJSON(t, resp, &scores)

	require.Len(t, scores, 2)
	// First-to-score ranks higher (earlier achieved_at wins)
	assert.Equal(t, "first", scores[0].UserID)
	assert.Equal(t, "second", scores[1].UserID)
}

func TestListScores_LimitsResults(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	for i := range 5 {
		submitScore(t, boardID, fmt.Sprintf("player_%d", i), (i+1)*100)
	}

	path := fmt.Sprintf("%s/%s/scores?n=3", boardsPath, boardID)
	resp := performRequest(t, http.MethodGet, path, nil)

	var scores []scoreResponse
	parseJSON(t, resp, &scores)

	assert.Len(t, scores, 3)
	// Top 3 should be 500, 400, 300
	assert.Equal(t, 500, scores[0].Score)
	assert.Equal(t, 400, scores[1].Score)
	assert.Equal(t, 300, scores[2].Score)
}

func TestListScores_DefaultN(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	submitScore(t, boardID, "player1", 100)

	// No n parameter — should default to 10
	path := fmt.Sprintf("%s/%s/scores", boardsPath, boardID)
	resp := performRequest(t, http.MethodGet, path, nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestListScores_InvalidLimit(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	path := fmt.Sprintf("%s/%s/scores?n=-5", boardsPath, boardID)
	resp := performRequest(t, http.MethodGet, path, nil)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestListScores_NonNumericN(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	path := fmt.Sprintf("%s/%s/scores?n=abc", boardsPath, boardID)
	resp := performRequest(t, http.MethodGet, path, nil)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestListScores_BoardNotFound(t *testing.T) {
	cleanBoards(t)

	fakeID := "123e4567-e89b-12d3-a456-426614174000"
	path := fmt.Sprintf("%s/%s/scores?n=10", boardsPath, fakeID)
	resp := performRequest(t, http.MethodGet, path, nil)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestListScores_BoardIsolation(t *testing.T) {
	cleanBoards(t)

	boardA := createDummyBoard(t)
	boardB := createDummyBoard(t)

	submitScore(t, boardA, "playerA", 1000)
	submitScore(t, boardB, "playerB", 2000)

	// Board A should only have playerA
	pathA := fmt.Sprintf("%s/%s/scores?n=10", boardsPath, boardA)
	respA := performRequest(t, http.MethodGet, pathA, nil)
	var scoresA []scoreResponse
	parseJSON(t, respA, &scoresA)

	require.Len(t, scoresA, 1)
	assert.Equal(t, "playerA", scoresA[0].UserID)

	// Board B should only have playerB
	pathB := fmt.Sprintf("%s/%s/scores?n=10", boardsPath, boardB)
	respB := performRequest(t, http.MethodGet, pathB, nil)
	var scoresB []scoreResponse
	parseJSON(t, respB, &scoresB)

	require.Len(t, scoresB, 1)
	assert.Equal(t, "playerB", scoresB[0].UserID)
}

// --- GET /api/v1/boards/:boardId/scores/:userId/surroundings ---

func TestSurroundings_Success(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	submitScore(t, boardID, "top", 1000)
	time.Sleep(5 * time.Millisecond)
	submitScore(t, boardID, "mid", 500)
	time.Sleep(5 * time.Millisecond)
	submitScore(t, boardID, "bottom", 100)

	path := fmt.Sprintf("%s/%s/scores/mid/surroundings?n=5", boardsPath, boardID)
	resp := performRequest(t, http.MethodGet, path, nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result surroundingsResponse
	parseJSON(t, resp, &result)

	assert.Equal(t, "mid", result.User.UserID)
	assert.Equal(t, 500, result.User.Score)

	require.Len(t, result.Above, 1)
	assert.Equal(t, "top", result.Above[0].UserID)

	require.Len(t, result.Below, 1)
	assert.Equal(t, "bottom", result.Below[0].UserID)
}

func TestSurroundings_TopPlayer(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	submitScore(t, boardID, "top", 1000)
	time.Sleep(5 * time.Millisecond)
	submitScore(t, boardID, "second", 500)

	path := fmt.Sprintf("%s/%s/scores/top/surroundings?n=5", boardsPath, boardID)
	resp := performRequest(t, http.MethodGet, path, nil)

	var result surroundingsResponse
	parseJSON(t, resp, &result)

	assert.Equal(t, "top", result.User.UserID)
	assert.Empty(t, result.Above, "top player should have no one above")
	require.Len(t, result.Below, 1)
	assert.Equal(t, "second", result.Below[0].UserID)
}

func TestSurroundings_BottomPlayer(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	submitScore(t, boardID, "first", 1000)
	time.Sleep(5 * time.Millisecond)
	submitScore(t, boardID, "last", 100)

	path := fmt.Sprintf("%s/%s/scores/last/surroundings?n=5", boardsPath, boardID)
	resp := performRequest(t, http.MethodGet, path, nil)

	var result surroundingsResponse
	parseJSON(t, resp, &result)

	assert.Equal(t, "last", result.User.UserID)
	require.Len(t, result.Above, 1)
	assert.Equal(t, "first", result.Above[0].UserID)
	assert.Empty(t, result.Below, "bottom player should have no one below")
}

func TestSurroundings_UserNotFound(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	path := fmt.Sprintf("%s/%s/scores/nonexistent/surroundings?n=5", boardsPath, boardID)
	resp := performRequest(t, http.MethodGet, path, nil)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSurroundings_BoardNotFound(t *testing.T) {
	cleanBoards(t)

	fakeID := "123e4567-e89b-12d3-a456-426614174000"
	path := fmt.Sprintf("%s/%s/scores/player1/surroundings?n=5", boardsPath, fakeID)
	resp := performRequest(t, http.MethodGet, path, nil)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSurroundings_InvalidLimit(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	path := fmt.Sprintf("%s/%s/scores/testUser/surroundings?n=0", boardsPath, boardID)
	resp := performRequest(t, http.MethodGet, path, nil)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSurroundings_LimitsAboveAndBelow(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	// Create 7 players: scores 100..700
	for i := 1; i <= 7; i++ {
		submitScore(t, boardID, fmt.Sprintf("p%d", i), i*100)
		time.Sleep(5 * time.Millisecond)
	}

	// Ask for surroundings of p4 (score 400) with n=2
	path := fmt.Sprintf("%s/%s/scores/p4/surroundings?n=2", boardsPath, boardID)
	resp := performRequest(t, http.MethodGet, path, nil)

	var result surroundingsResponse
	parseJSON(t, resp, &result)

	assert.Equal(t, "p4", result.User.UserID)
	assert.Equal(t, 400, result.User.Score)

	// Above: p6(600) and p5(500) — ordered highest first after reverse
	require.Len(t, result.Above, 2)
	assert.Equal(t, "p6", result.Above[0].UserID)
	assert.Equal(t, "p5", result.Above[1].UserID)

	// Below: p3(300) and p2(200) — ordered closest to pivot first
	require.Len(t, result.Below, 2)
	assert.Equal(t, "p3", result.Below[0].UserID)
	assert.Equal(t, "p2", result.Below[1].UserID)
}

// --- POST /api/v1/boards/:boardId/scores/seed ---

func TestSeedScores_Success(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	body := map[string]any{
		"n": 5,
	}
	path := fmt.Sprintf("%s/%s/scores/seed", boardsPath, boardID)
	resp := performRequest(t, http.MethodPost, path, body)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result seedResponse
	parseJSON(t, resp, &result)
	assert.Equal(t, 5, result.ScoresCreated)

	// Verify scores were actually created
	scoresPath := fmt.Sprintf("%s/%s/scores?n=100", boardsPath, boardID)
	scoresResp := performRequest(t, http.MethodGet, scoresPath, nil)
	var scores []scoreResponse
	parseJSON(t, scoresResp, &scores)
	assert.Len(t, scores, 5)
}

func TestSeedScores_InvalidCount(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	body := map[string]any{
		"n": -1,
	}
	path := fmt.Sprintf("%s/%s/scores/seed", boardsPath, boardID)
	resp := performRequest(t, http.MethodPost, path, body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSeedScores_ZeroCount(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	body := map[string]any{
		"n": 0,
	}
	path := fmt.Sprintf("%s/%s/scores/seed", boardsPath, boardID)
	resp := performRequest(t, http.MethodPost, path, body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSeedScores_BoardNotFound(t *testing.T) {
	cleanBoards(t)

	fakeID := "123e4567-e89b-12d3-a456-426614174000"
	body := map[string]any{
		"n": 5,
	}
	path := fmt.Sprintf("%s/%s/scores/seed", boardsPath, fakeID)
	resp := performRequest(t, http.MethodPost, path, body)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
