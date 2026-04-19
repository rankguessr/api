package service

import (
	"context"
	"time"

	"github.com/rankguessr/api/internal/repo"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/ranking"
)

type Guess interface {
	CountFromDate(ctx context.Context, from time.Time) (int, error)

	FindLatest(ctx context.Context) ([]domain.Guess, error)
	FindByUser(ctx context.Context, userId, limit int) ([]domain.Guess, error)
	FindById(ctx context.Context, id string) (domain.Guess, error)
	FindTopFromDate(ctx context.Context, from time.Time, limit int) ([]domain.GuessExtended, error)

	Create(ctx context.Context, userId int, input domain.GuessCreate) (int, domain.Guess, error)
}

type guess struct {
	repo repo.Guesses
}

func NewGuess(repo repo.Guesses) Guess {
	return &guess{repo: repo}
}

func (g *guess) FindById(ctx context.Context, id string) (domain.Guess, error) {
	return g.repo.FindById(ctx, id)
}

func (g *guess) CountFromDate(ctx context.Context, from time.Time) (int, error) {
	return g.repo.CountFromDate(ctx, from)
}

func (g *guess) FindLatest(ctx context.Context) ([]domain.Guess, error) {
	return g.repo.FindLatest(ctx)
}

func (g *guess) FindTopFromDate(ctx context.Context, from time.Time, limit int) ([]domain.GuessExtended, error) {
	return g.repo.FindTopFromDate(ctx, from, limit)
}

func (g *guess) Create(ctx context.Context, userId int, input domain.GuessCreate) (int, domain.Guess, error) {
	elo, err := ranking.Calculate(input.Guess, input.ActualRank)
	if err != nil {
		return 0, domain.Guess{}, err
	}

	return g.repo.Create(ctx, userId, elo, input)
}

func (g *guess) FindByUser(ctx context.Context, userId, limit int) ([]domain.Guess, error) {
	return g.repo.FindByUser(ctx, userId, limit)
}
