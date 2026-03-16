package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a board purely for testing scores
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

func TestSubmitScore_NegativeScore(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	body := map[string]any{
		"userId": "player2",
		"score":  -50, // Validation tag gt=0 specifies greater than 0
	}
	path := fmt.Sprintf("%s/%s/scores", boardsPath, boardID)
	resp := performRequest(t, http.MethodPost, path, body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// --- GET /api/v1/boards/:boardId/scores ---

func TestListScores_Success(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	// Submit a score so the list isn't empty, though empty is valid too
	body := map[string]any{
		"userId": "player1",
		"score":  1000,
	}
	path := fmt.Sprintf("%s/%s/scores", boardsPath, boardID)
	performRequest(t, http.MethodPost, path, body)

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

// --- GET /api/v1/boards/:boardId/scores/:userId/surroundings ---

func TestSurroundings_Success(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	// Must create the user first to ask for their surroundings (otherwise 404)
	body := map[string]any{
		"userId": "testUser",
		"score":  1000,
	}
	submitPath := fmt.Sprintf("%s/%s/scores", boardsPath, boardID)
	performRequest(t, http.MethodPost, submitPath, body)

	path := fmt.Sprintf("%s/%s/scores/testUser/surroundings", boardsPath, boardID)
	resp := performRequest(t, http.MethodGet, path, nil)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSurroundings_InvalidLimit(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	path := fmt.Sprintf("%s/%s/scores/testUser/surroundings?n=0", boardsPath, boardID)
	resp := performRequest(t, http.MethodGet, path, nil)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
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
}

func TestSeedScores_InvalidCount(t *testing.T) {
	cleanBoards(t)
	boardID := createDummyBoard(t)

	body := map[string]any{
		"n": -1, // must be > 0
	}
	path := fmt.Sprintf("%s/%s/scores/seed", boardsPath, boardID)
	resp := performRequest(t, http.MethodPost, path, body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
