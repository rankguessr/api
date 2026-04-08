package repo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/utils"
)

type Rooms interface {
	FindByID(ctx context.Context, id string) (domain.Room, error)
	SetClosed(ctx context.Context, id string, value bool) error
	Create(ctx context.Context, playerId, userId, scoreId int) (domain.Room, error)
}

type rooms struct {
	pool *pgxpool.Pool
}

var rowToRoom = pgx.RowToStructByName[domain.Room]

func NewRooms(pool *pgxpool.Pool) Rooms {
	return &rooms{pool: pool}
}

func (s *rooms) SetClosed(ctx context.Context, id string, value bool) error {
	_, err := s.pool.Exec(ctx, "UPDATE rooms SET is_closed = $1 WHERE id = $2", value, id)
	if err != nil {
		return err
	}

	return nil
}

const createRoomQuery = `
	INSERT INTO rooms (id, player_id, user_id, score_id) 
	VALUES (@id, @playerId, @userId, @scoreId) RETURNING *
`

func (s *rooms) Create(ctx context.Context, playerId, userId, scoreId int) (domain.Room, error) {
	rows, err := s.pool.Query(ctx, createRoomQuery, pgx.NamedArgs{
		"userId":   userId,
		"scoreId":  scoreId,
		"playerId": playerId,
		"id":       utils.NewID(),
	})
	if err != nil {
		return domain.Room{}, err
	}

	return pgx.CollectOneRow(rows, rowToRoom)
}

const findRoomByIDQuery = `
	SELECT *
	FROM rooms
	WHERE id = $1
`

func (s *rooms) FindByID(ctx context.Context, id string) (domain.Room, error) {
	rows, err := s.pool.Query(ctx, findRoomByIDQuery, id)
	if err != nil {
		return domain.Room{}, err
	}

	return pgx.CollectOneRow(rows, rowToRoom)
}
