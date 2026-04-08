package domain

import "time"

type SessionBase struct {
	ID           string    `json:"id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type Session struct {
	UserID int `json:"user_id"`
	SessionBase
}

type SessionExtended struct {
	User User `json:"user"`
	SessionBase
}
