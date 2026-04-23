package domain

import "time"

type GuessCreate struct {
	PlayerID     int      `json:"player_id"`
	Guess        int      `json:"guess"`
	ActualRank   int      `json:"actual_rank"`
	ScoreID      int      `json:"score_id"`
	BeatmapID    int      `json:"beatmap_id"`
	Kind         RoomKind `json:"kind"`
	BeatmapSetID int      `json:"beatmapset_id"`
}

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

type RefillResult struct {
	AvailableGuesses uint      `json:"available_guesses"`
	RefilledAt       time.Time `json:"refilled_at"`
}
