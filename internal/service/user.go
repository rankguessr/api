package service

import (
	"context"

	"github.com/rankguessr/api/internal/repo"
	"github.com/rankguessr/api/pkg/domain"
)

type User interface {
	Upsert(ctx context.Context, osuId int, username, avatarURL, countryCode string) error

	FindByOsuID(ctx context.Context, osuId int) (domain.User, error)
}

type user struct {
	repo repo.Users
}

func NewUser(repo repo.Users) User {
	return &user{repo: repo}
}

func (u *user) FindByOsuID(ctx context.Context, osuId int) (domain.User, error) {
	return u.repo.FindByOsuID(ctx, osuId)
}

func (u *user) Upsert(ctx context.Context, osuId int, username, avatarURL, countryCode string) error {
	return u.repo.Upsert(ctx, osuId, username, avatarURL, countryCode)
}
