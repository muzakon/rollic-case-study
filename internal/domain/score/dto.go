package score

// ScoreResponse is the response DTO for a single score entry.
type ScoreResponse struct {
	UserID string `json:"userId"`
	Score  int    `json:"score"`
}

// SubmitRequest is the request DTO for submitting a score.
type SubmitRequest struct {
	UserID string `json:"userId" validate:"required"`
	Score  int    `json:"score" validate:"required,gt=0"`
}

// SubmitResponse is the response DTO for the submit endpoint.
type SubmitResponse struct {
	BoardID string `json:"boardId"`
	UserID  string `json:"userId"`
	Score   int    `json:"score"`
}

// SeedRequest is the request DTO for the seed endpoint.
type SeedRequest struct {
	N int `json:"n" validate:"required,gt=0"`
}

// SeedResponse is the response DTO for the seed endpoint.
type SeedResponse struct {
	ScoresCreated int `json:"scoresCreated"`
}

// SurroundingsResponse is the response DTO for the surroundings endpoint.
type SurroundingsResponse struct {
	User  ScoreResponse   `json:"user"`
	Above []ScoreResponse `json:"above"`
	Below []ScoreResponse `json:"below"`
}
