package score

// ScoreResponse is the response DTO for a single score entry.
type ScoreResponse struct {
	UserID string `json:"userId"`
	Score  int    `json:"score"`
}
