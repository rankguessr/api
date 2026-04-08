package domain

import "time"

type Guess struct {
	ID         string    `json:"id"`
	UserID     int       `json:"user_id"`
	PlayerID   int       `json:"player_id"`
	Guess      int       `json:"guess"`
	ActualRank int       `json:"actual_rank"`
	Elo        int       `json:"elo"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
