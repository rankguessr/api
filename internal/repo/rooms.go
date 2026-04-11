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
	FindByUser(ctx context.Context, userId int) (domain.Room, error)
	Create(ctx context.Context, playerId, userId, scoreId int) (domain.Room, error)
	DeleteById(ctx context.Context, id string) error

	UpdateGuessID(ctx context.Context, id string, guessId string) error
	UpdateScore(ctx context.Context, id string, playerId, scoreId int, guessId *string) (domain.Room, error)
}

type rooms struct {
	pool *pgxpool.Pool
}

var rowToRoom = pgx.RowToStructByName[domain.Room]

func NewRooms(pool *pgxpool.Pool) Rooms {
	return &rooms{pool: pool}
}

func (s *rooms) DeleteById(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM rooms WHERE id = $1
	`, id)
	return err
}

func (s *rooms) FindByUser(ctx context.Context, userId int) (domain.Room, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT * FROM rooms 
		WHERE user_id = $1 AND guess_id IS NULL
		ORDER BY created_at DESC LIMIT 1
	`, userId)
	if err != nil {
		return domain.Room{}, err
	}

	return pgx.CollectOneRow(rows, rowToRoom)
}

func (s *rooms) UpdateGuessID(ctx context.Context, id string, guessId string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE rooms SET guess_id = $1 WHERE id = $2
	`, guessId, id)
	return err
}

func (s *rooms) UpdateScore(ctx context.Context, id string, playerId, scoreId int, guessId *string) (domain.Room, error) {
	rows, err := s.pool.Query(ctx, `
		UPDATE rooms SET 
			player_id = $1, 
			score_id = $2, 
			guess_id = $3 
		WHERE id = $4 
		RETURNING *
	`, playerId, scoreId, guessId, id)
	if err != nil {
		return domain.Room{}, err
	}

	return pgx.CollectOneRow(rows, rowToRoom)
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
