package domain

import (
	"time"
)

type User struct {
	Elo              uint      `json:"elo"`
	OsuID            int       `json:"osu_id"`
	Username         string    `json:"username"`
	AvatarURL        string    `json:"avatar_url"`
	CountryCode      string    `json:"country_code"`
	IsAdmin          bool      `json:"is_admin"`
	AvailableGuesses uint      `json:"available_guesses"`
	RefilledAt       time.Time `json:"refilled_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type UserExtended struct {
	User
	TotalGuesses uint `json:"total_guesses"`
	Rank         uint `json:"rank"`
}
