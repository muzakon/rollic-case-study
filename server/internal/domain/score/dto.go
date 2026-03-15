package score

// ScoreResponse is the response DTO for a single score entry.
type ScoreResponse struct {
	UserID string `json:"userId"`
	Score  int    `json:"score"`
}

// SurroundingsResponse is the response DTO for the surroundings endpoint.
type SurroundingsResponse struct {
	User  ScoreResponse   `json:"user"`
	Above []ScoreResponse `json:"above"`
	Below []ScoreResponse `json:"below"`
}
