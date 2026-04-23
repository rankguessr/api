package domain

import (
	"time"

	"github.com/rankguessr/api/pkg/osuapi"
)

type RoomKind string

const (
	RoomKindRankedV1     RoomKind = "v1"
	RoomKindRankedV2     RoomKind = "v2"
	RoomKindSubmissionV2 RoomKind = "v2sub"
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

type RoomExtended struct {
	Room
	Score osuapi.Score `json:"score"`
}
