package repo

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rankguessr/api/pkg/domain"
)

var rowToPlayer = pgx.RowToStructByName[domain.Player]

type Players interface {
	FindRandom(ctx context.Context) (domain.Player, error)
	CreateMany(ctx context.Context, ids []domain.PlayerCreate) (int64, error)
}

type players struct {
	pool *pgxpool.Pool
}

func NewPlayers(pool *pgxpool.Pool) Players {
	return &players{pool: pool}
}

const findRandomQuery = `
	SELECT *
	FROM players
	ORDER BY RANDOM()
	LIMIT 1
`

func (p *players) FindRandom(ctx context.Context) (domain.Player, error) {
	rows, err := p.pool.Query(ctx, findRandomQuery)
	if err != nil {
		return domain.Player{}, err
	}

	return pgx.CollectExactlyOneRow(rows, rowToPlayer)
}

const createManyQuery = `
	INSERT INTO players (osu_id, source)
	VALUES (@osu_id, @source)
	ON CONFLICT (osu_id) DO NOTHING
`

func (p *players) CreateMany(ctx context.Context, players []domain.PlayerCreate) (int64, error) {
	total := int64(0)
	batch := &pgx.Batch{}

	for _, p := range players {
		batch.Queue(createManyQuery, pgx.NamedArgs{
			"source": p.Source,
			"osu_id": p.OsuId,
		})
	}

	res := p.pool.SendBatch(ctx, batch)
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
