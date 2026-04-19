package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/utils"
)

const maxGuesses = 15

type Rooms interface {
	FindByID(ctx context.Context, id string) (domain.Room, error)
	FindByUser(ctx context.Context, userId int) (domain.Room, error)
	FindByUserUnguessed(ctx context.Context, userId int) (domain.Room, error)

	Create(ctx context.Context, playerId, userId, scoreId int) (domain.Room, error)

	DeleteById(ctx context.Context, id string) error
	DeleteByUserUnguessed(ctx context.Context, userId int) error
	DeleteByUser(ctx context.Context, userId int) error

	RefillForUser(ctx context.Context, userId int, sub uint) (domain.RefillResult, error)

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

func (r *rooms) RefillForUser(ctx context.Context, osuId int, sub uint) (domain.RefillResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.RefillResult{}, err
	}

	rows, err := tx.Query(ctx, `
		SELECT * FROM users 
		WHERE osu_id = $1
		FOR UPDATE
	`, osuId)
	if err != nil {
		return domain.RefillResult{}, err
	}

	user, err := pgx.CollectOneRow(rows, rowToUser)
	if err != nil {
		tx.Rollback(ctx)
		return domain.RefillResult{}, err
	}

	now := time.Now()

	rem := (now.Sub(user.RefilledAt)) % refillInterval
	refilled := min(maxGuesses, user.AvailableGuesses+uint((now.Sub(user.RefilledAt))/refillInterval))
	if refilled < sub {
		tx.Rollback(ctx)
		return domain.RefillResult{}, utils.ErrNotEnoughGuesses
	}

	newGuesses := refilled - sub

	_, err = tx.Exec(ctx, `
		UPDATE users 
		SET available_guesses = $1, refilled_at = $2
		WHERE osu_id = $3
	`, newGuesses, now.Add(-rem), osuId)
	if err != nil {
		tx.Rollback(ctx)
		return domain.RefillResult{}, err
	}

	return domain.RefillResult{
		RefilledAt:       now.Add(-rem),
		AvailableGuesses: newGuesses,
	}, tx.Commit(ctx)
}

func (s *rooms) DeleteByUserUnguessed(ctx context.Context, userId int) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM rooms WHERE user_id = $1 AND guess_id IS NULL
	`, userId)
	return err
}

func (s *rooms) DeleteByUser(ctx context.Context, userId int) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM rooms WHERE user_id = $1 AND guess_id IS NOT NULL
	`, userId)
	return err
}

func (s *rooms) FindByUserUnguessed(ctx context.Context, userId int) (domain.Room, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT * FROM rooms 
		WHERE user_id = $1 AND guess_id IS NULL
	`, userId)
	if err != nil {
		return domain.Room{}, err
	}

	return pgx.CollectOneRow(rows, rowToRoom)
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
