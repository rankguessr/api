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
var rowToGuessExtended = pgx.RowToStructByName[domain.GuessExtended]

type Guesses interface {
	CountFromDate(ctx context.Context, from time.Time) (int, error)

	FindLatest(ctx context.Context) ([]domain.Guess, error)
	FindById(ctx context.Context, id string) (domain.Guess, error)
	FindByUser(ctx context.Context, userId, limit int) ([]domain.Guess, error)
	FindTopFromDate(ctx context.Context, from time.Time, limit int) ([]domain.GuessExtended, error)

	Create(ctx context.Context, userId, playerId, guess, actualRank, elo, scoreId, beatmapId, beatmapSetId int) (int, domain.Guess, error)
}

type guesses struct {
	pool *pgxpool.Pool
}

func NewGuesses(pool *pgxpool.Pool) Guesses {
	return &guesses{pool: pool}
}

func (g *guesses) FindById(ctx context.Context, id string) (domain.Guess, error) {
	rows, err := g.pool.Query(ctx, "SELECT * FROM guesses WHERE id = $1", id)
	if err != nil {
		return domain.Guess{}, err
	}

	return pgx.CollectOneRow(rows, rowToGuess)
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
	SELECT 
		g.*, to_json(u) AS user 
	FROM guesses g
	JOIN users u ON g.user_id = u.osu_id
	WHERE g.created_at >= $1
	ORDER BY g.elo DESC 
	LIMIT $2
`

func (g *guesses) FindTopFromDate(ctx context.Context, from time.Time, limit int) ([]domain.GuessExtended, error) {
	rows, err := g.pool.Query(ctx, findTopFromDateQuery, from, limit)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToFunc[domain.GuessExtended](rowToGuessExtended))
}

const createGuessQuery = `
	INSERT INTO guesses (id, user_id, player_id, guess, actual_rank, elo, beatmap_id, beatmapset_id, score_id)
	VALUES (@id, @userId, @playerId, @guess, @actualRank, @elo, @beatmapId, @beatmapSetId, @scoreId) RETURNING *
`

// returns new elo & created guess
func (g *guesses) Create(ctx context.Context, userId, playerId, guess, actualRank, elo, scoreId, beatmapId, beatmapSetId int) (int, domain.Guess, error) {
	tx, err := g.pool.Begin(ctx)
	if err != nil {
		return 0, domain.Guess{}, err
	}

	rows, err := tx.Query(ctx, createGuessQuery, pgx.NamedArgs{
		"elo":        elo,
		"guess":      guess,
		"userId":     userId,
		"playerId":   playerId,
		"actualRank": actualRank,
		"id":         utils.NewID(),

		"scoreId":      scoreId,
		"beatmapId":    beatmapId,
		"beatmapSetId": beatmapSetId,
	})
	if err != nil {
		tx.Rollback(ctx)
		return 0, domain.Guess{}, err
	}

	guessRes, err := pgx.CollectOneRow(rows, rowToGuess)
	if err != nil {
		tx.Rollback(ctx)
		return 0, domain.Guess{}, err
	}

	rows, err = tx.Query(ctx, "UPDATE users SET elo = GREATEST(0, elo + $1) WHERE osu_id = $2 RETURNING elo", elo, userId)
	if err != nil {
		tx.Rollback(ctx)
		return 0, domain.Guess{}, err
	}

	newElo, err := pgx.CollectOneRow(rows, pgx.RowTo[int])

	return newElo, guessRes, tx.Commit(ctx)
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
