package domain

import "time"

type Guess struct {
	ID           string    `json:"id"`
	UserID       int       `json:"user_id"`
	PlayerID     int       `json:"player_id"`
	Guess        int       `json:"guess"`
	ActualRank   int       `json:"actual_rank"`
	Elo          int       `json:"elo"`
	ScoreID      int       `json:"score_id"`
	Kind         RoomKind  `json:"kind"`
	BeatmapID    int       `json:"beatmap_id"`
	BeatmapSetID int       `json:"beatmapset_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type GuessExtended struct {
	Guess
	User User `json:"user"`
}
