package domain

type Submission struct {
	SubmissionCreate
	ID         string `json:"id"`
	IsAccepted bool   `json:"is_accepted"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type SubmissionCreate struct {
	UserID       int    `json:"user_id"`
	PlayerID     int    `json:"player_id"`
	ScoreID      int    `json:"score_id"`
	BeatmapID    int    `json:"beatmap_id"`
	BeatmapsetID int    `json:"beatmapset_id"`
	Comment      string `json:"comment"`
}
