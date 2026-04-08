package domain

import (
	"time"
)

type User struct {
	OsuID       int       `json:"osu_id"`
	Username    string    `json:"username"`
	AvatarURL   string    `json:"avatar_url"`
	CountryCode string    `json:"country_code"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
