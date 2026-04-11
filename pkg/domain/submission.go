package domain

type Submission struct {
	ID         string `json:"id"`
	UserID     int    `json:"user_id"`
	PlayerID   int    `json:"player_id"`
	ScoreID    int    `json:"score_id"`
	IsAccepted bool   `json:"is_accepted"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}
