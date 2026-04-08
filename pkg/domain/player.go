package domain

import "time"

type PlayerCreate struct {
	OsuId  int    `json:"osu_id"`
	Source string `json:"source"`
}

type Player struct {
	PlayerCreate
	CreatedAt time.Time `json:"created_at"`
	CheckedAt time.Time `json:"checked_at"`
}
