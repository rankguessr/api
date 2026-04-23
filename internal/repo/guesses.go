package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rankguessr/api/internal/uow"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/utils"
)

var (
	rowToGuess         = pgx.RowToStructByName[domain.Guess]
	rowToGuessExtended = pgx.RowToStructByName[domain.GuessExtended]
)

type Guesses interface {
	CountFromDate(ctx context.Context, from time.Time) (int, error)

	FindLatest(ctx context.Context) ([]domain.Guess, error)
	FindById(ctx context.Context, id string) (domain.Guess, error)
	FindByUser(ctx context.Context, userId, limit int) ([]domain.Guess, error)
	FindTopFromDate(ctx context.Context, from time.Time, limit int) ([]domain.GuessExtended, error)

	Create(ctx context.Context, userId, elo int, input domain.GuessCreate) (domain.Guess, error)
}

type guesses struct {
	uow *uow.UnitOfWork
}

func NewGuesses(uow *uow.UnitOfWork) Guesses {
	return &guesses{uow: uow}
}

func (g *guesses) FindById(ctx context.Context, id string) (domain.Guess, error) {
	ex := g.uow.Executor(ctx)
	rows, err := ex.Query(ctx, "SELECT * FROM guesses WHERE id = $1", id)
	if err != nil {
		return domain.Guess{}, err
	}

	return pgx.CollectOneRow(rows, rowToGuess)
}

func (g *guesses) CountFromDate(ctx context.Context, from time.Time) (int, error) {
	ex := g.uow.Executor(ctx)
	rows, err := ex.Query(ctx, "SELECT COUNT(*) FROM guesses WHERE created_at >= $1 AND kind != 'v1'", from)
	if err != nil {
		return 0, err
	}

	return pgx.CollectExactlyOneRow(rows, pgx.RowTo[int])
}

const findLatestQuery = `
	SELECT * FROM guesses
	ORDER BY created_at DESC
	LIMIT 10
`

func (g *guesses) FindLatest(ctx context.Context) ([]domain.Guess, error) {
	ex := g.uow.Executor(ctx)
	rows, err := ex.Query(ctx, findLatestQuery)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, rowToGuess)
}

func (g *guesses) FindTopFromDate(ctx context.Context, from time.Time, limit int) ([]domain.GuessExtended, error) {
	query := `
		SELECT 
			g.*, to_json(u) AS user 
		FROM guesses g
		JOIN users u ON g.user_id = u.osu_id
		WHERE g.created_at >= $1 AND g.kind != 'v1'
		ORDER BY g.elo DESC 
		LIMIT $2
	`

	ex := g.uow.Executor(ctx)
	rows, err := ex.Query(ctx, query, from, limit)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, rowToGuessExtended)
}

func (g *guesses) Create(ctx context.Context, userId, elo int, input domain.GuessCreate) (domain.Guess, error) {
	ex := g.uow.Executor(ctx)
	rows, err := ex.Query(ctx, `
		INSERT INTO guesses (id, user_id, player_id, guess, actual_rank, elo, beatmap_id, beatmapset_id, score_id, kind)
		VALUES (@id, @userId, @playerId, @guess, @actualRank, @elo, @beatmapId, @beatmapSetId, @scoreId, @kind) RETURNING *
	`, pgx.NamedArgs{
		"elo":    elo,
		"userId": userId,
		"id":     utils.NewID(),

		"guess":        input.Guess,
		"kind":         input.Kind,
		"playerId":     input.PlayerID,
		"actualRank":   input.ActualRank,
		"scoreId":      input.ScoreID,
		"beatmapId":    input.BeatmapID,
		"beatmapSetId": input.BeatmapSetID,
	})
	if err != nil {
		return domain.Guess{}, err
	}

	return pgx.CollectOneRow(rows, rowToGuess)
}

func (g *guesses) FindByUser(ctx context.Context, userId, limit int) ([]domain.Guess, error) {
	ex := g.uow.Executor(ctx)
	rows, err := ex.Query(ctx, `
		SELECT * FROM guesses
		WHERE user_id = $1 AND kind != 'v1'
		ORDER BY created_at DESC
		LIMIT $2
	`, userId, limit)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, rowToGuess)
}
