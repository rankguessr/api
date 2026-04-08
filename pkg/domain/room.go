package domain

import "time"

type Room struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	ScoreID   int       `json:"score_id"`
	PlayerID  int       `json:"player_id"`
	IsClosed  bool      `json:"is_closed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
