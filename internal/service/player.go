package service

import (
	"context"

	"github.com/rankguessr/api/internal/repo"
	"github.com/rankguessr/api/pkg/domain"
)

type Players interface {
	FindRandom(ctx context.Context) (domain.Player, error)
	Find(ctx context.Context, limit, page int) ([]domain.Player, error)
	CreateMany(ctx context.Context, ids []domain.PlayerCreate) (int64, error)
}

type players struct {
	repo repo.Players
}

func NewPlayer(repo repo.Players) Players {
	return &players{repo: repo}
}

func (p *players) Find(ctx context.Context, limit, page int) ([]domain.Player, error) {
	return p.repo.Find(ctx, limit, page*limit)
}

func (p *players) FindRandom(ctx context.Context) (domain.Player, error) {
	return p.repo.FindRandom(ctx)
}

func (p *players) CreateMany(ctx context.Context, players []domain.PlayerCreate) (int64, error) {
	return p.repo.CreateMany(ctx, players)
}
