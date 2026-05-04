package repo

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/rankguessr/api/internal/uow"
	"github.com/rankguessr/api/pkg/domain"
)

var rowToPlayer = pgx.RowToStructByName[domain.Player]

type Players interface {
	Find(ctx context.Context, limit, offset int) ([]domain.Player, error)
	FindRandom(ctx context.Context) (domain.Player, error)
	CreateMany(ctx context.Context, ids []domain.PlayerCreate) (int64, error)
}

type players struct {
	uow *uow.UnitOfWork
}

func NewPlayers(uow *uow.UnitOfWork) Players {
	return &players{uow: uow}
}

func (p *players) Find(ctx context.Context, limit, offset int) ([]domain.Player, error) {
	ex := p.uow.Executor(ctx)
	rows, err := ex.Query(ctx, `
		SELECT *
		FROM players
		WHERE source = 'multi'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, rowToPlayer)
}

func (p *players) FindRandom(ctx context.Context) (domain.Player, error) {
	ex := p.uow.Executor(ctx)
	rows, err := ex.Query(ctx, "SELECT * FROM players ORDER BY RANDOM() LIMIT 1")
	if err != nil {
		return domain.Player{}, err
	}

	return pgx.CollectExactlyOneRow(rows, rowToPlayer)
}

func (p *players) CreateMany(ctx context.Context, players []domain.PlayerCreate) (int64, error) {
	total := int64(0)
	batch := &pgx.Batch{}
	ex := p.uow.Executor(ctx)

	for _, p := range players {
		batch.Queue(`
			INSERT INTO players (osu_id, source)
			VALUES (@osu_id, @source)
			ON CONFLICT (osu_id) DO NOTHING
		`, pgx.NamedArgs{
			"source": p.Source,
			"osu_id": p.OsuId,
		})
	}

	res := ex.SendBatch(ctx, batch)
	defer res.Close()

	for range len(players) {
		tag, err := res.Exec()
		if err != nil {
			log.Println("Error occurred while creating player:", err)
		}

		total += tag.RowsAffected()
	}

	return total, nil
}
