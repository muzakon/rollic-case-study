package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// performRequest sends an HTTP request through Fiber's app.Test() without starting a real server.
// If body is non-nil, it is marshaled to JSON and set as the request body.
func performRequest(t *testing.T, method, path string, body any) *http.Response {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		require.NoError(t, err, "failed to marshal request body")
		reqBody = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(t, err, "failed to create request")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := testApp.Test(req)
	require.NoError(t, err, "app.Test failed")

	return resp
}

// parseJSON reads the response body and unmarshals it into the given target.
func parseJSON(t *testing.T, resp *http.Response, target any) {
	t.Helper()

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read response body")

	err = json.Unmarshal(bodyBytes, target)
	require.NoError(t, err, "failed to unmarshal response body: %s", string(bodyBytes))
}

// readBody reads and returns the raw response body as a string.
func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read response body")
	return string(bodyBytes)
}

// cleanBoards truncates the boards table (cascading to scores) to ensure test isolation.
func cleanBoards(t *testing.T) {
	t.Helper()
	err := testDB.Exec("TRUNCATE TABLE boards CASCADE").Error
	require.NoError(t, err, "failed to truncate boards table")
}
