package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/utils"
)

var rowToGuess = pgx.RowToStructByName[domain.Guess]

type Guesses interface {
	CountFromDate(ctx context.Context, from time.Time) (int, error)

	FindLatest(ctx context.Context) ([]domain.Guess, error)
	FindByUser(ctx context.Context, userId, limit int) ([]domain.Guess, error)
	FindTopFromDate(ctx context.Context, from time.Time, limit int) ([]domain.Guess, error)

	Create(ctx context.Context, userId, playerId, guess, actualRank, elo int) (domain.Guess, error)
}

type guesses struct {
	pool *pgxpool.Pool
}

func NewGuesses(pool *pgxpool.Pool) Guesses {
	return &guesses{pool: pool}
}

const countFromDateQuery = `
	SELECT COUNT(*) FROM guesses
	WHERE created_at >= $1
`

func (g *guesses) CountFromDate(ctx context.Context, from time.Time) (int, error) {
	var count int
	err := g.pool.QueryRow(ctx, countFromDateQuery, from).Scan(&count)
	return count, err
}

const findLatestQuery = `
	SELECT * FROM guesses
	ORDER BY created_at DESC
	LIMIT 10
`

func (g *guesses) FindLatest(ctx context.Context) ([]domain.Guess, error) {
	rows, err := g.pool.Query(ctx, findLatestQuery)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, rowToGuess)
}

const findTopFromDateQuery = `
	SELECT * FROM guesses
	WHERE created_at >= $1
	ORDER BY elo DESC 
	LIMIT $2
`

func (g *guesses) FindTopFromDate(ctx context.Context, from time.Time, limit int) ([]domain.Guess, error) {
	rows, err := g.pool.Query(ctx, findTopFromDateQuery, from, limit)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, rowToGuess)
}

const createGuessQuery = `
	INSERT INTO guesses (id, user_id, player_id, guess, actual_rank, elo)
	VALUES (@id, @userId, @playerId, @guess, @actualRank, @elo) RETURNING *
`

func (g *guesses) Create(ctx context.Context, userId int, playerId int, guess int, actualRank int, elo int) (domain.Guess, error) {
	rows, err := g.pool.Query(ctx, createGuessQuery, pgx.NamedArgs{
		"elo":        elo,
		"guess":      guess,
		"userId":     userId,
		"playerId":   playerId,
		"actualRank": actualRank,
		"id":         utils.NewID(),
	})
	if err != nil {
		return domain.Guess{}, err
	}

	return pgx.CollectOneRow(rows, rowToGuess)
}

func (g *guesses) FindByUser(ctx context.Context, userId, limit int) ([]domain.Guess, error) {
	rows, err := g.pool.Query(ctx, `
		SELECT * FROM guesses
		WHERE user_id = $1 
		ORDER BY created_at DESC
		LIMIT $2
	`, userId, limit)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, rowToGuess)
}
