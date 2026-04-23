package service

import (
	"context"
	"time"

	"github.com/rankguessr/api/internal/repo"
	"github.com/rankguessr/api/internal/uow"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/ranking"
)

type Guess interface {
	CountFromDate(ctx context.Context, from time.Time) (int, error)

	FindLatest(ctx context.Context) ([]domain.Guess, error)
	FindByUser(ctx context.Context, userId, limit int) ([]domain.Guess, error)
	FindById(ctx context.Context, id string) (domain.Guess, error)
	FindTopFromDate(ctx context.Context, from time.Time, limit int) ([]domain.GuessExtended, error)

	CreateAndUpdateUserElo(ctx context.Context, userId int, input domain.GuessCreate) (int, domain.Guess, error)
}

type guess struct {
	guesses repo.Guesses
	users   repo.Users
	uow     *uow.UnitOfWork
}

func NewGuess(guesses repo.Guesses, users repo.Users, uow *uow.UnitOfWork) Guess {
	return &guess{guesses: guesses, users: users, uow: uow}
}

func (g *guess) FindById(ctx context.Context, id string) (domain.Guess, error) {
	return g.guesses.FindById(ctx, id)
}

func (g *guess) CountFromDate(ctx context.Context, from time.Time) (int, error) {
	return g.guesses.CountFromDate(ctx, from)
}

func (g *guess) FindLatest(ctx context.Context) ([]domain.Guess, error) {
	return g.guesses.FindLatest(ctx)
}

func (g *guess) FindTopFromDate(ctx context.Context, from time.Time, limit int) ([]domain.GuessExtended, error) {
	return g.guesses.FindTopFromDate(ctx, from, limit)
}

func (g *guess) CreateAndUpdateUserElo(ctx context.Context, userId int, input domain.GuessCreate) (newElo int, guess domain.Guess, err error) {
	elo, err := ranking.Calculate(input.Guess, input.ActualRank)
	if err != nil {
		return 0, domain.Guess{}, err
	}

	err = g.uow.WithTx(ctx, func(ctx context.Context) error {
		guess, err = g.guesses.Create(ctx, userId, elo, input)
		if err != nil {
			return err
		}

		newElo, err = g.users.UpdateElo(ctx, userId, elo)
		if err != nil {
			return err
		}

		return nil
	})

	return newElo, guess, err
}

func (g *guess) FindByUser(ctx context.Context, userId, limit int) ([]domain.Guess, error) {
	return g.guesses.FindByUser(ctx, userId, limit)
}
