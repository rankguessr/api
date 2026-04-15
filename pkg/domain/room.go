package domain

import "time"

type RoomKind string

const (
	RoomKindRanked     RoomKind = "ranked"
	RoomKindSubmission RoomKind = "submission"
)

type Room struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	ScoreID   int       `json:"score_id"`
	PlayerID  int       `json:"player_id"`
	Kind      RoomKind  `json:"kind"`
	GuessID   *string   `json:"guess_id" db:"guess_id"`
	ClosesAt  time.Time `json:"closes_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
