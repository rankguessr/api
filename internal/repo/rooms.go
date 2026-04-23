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
	rowToRoom = pgx.RowToStructByName[domain.Room]
)

type Rooms interface {
	FindByID(ctx context.Context, id string) (domain.Room, error)
	FindByUser(ctx context.Context, userId int) (domain.Room, error)
	FindByUserUnguessed(ctx context.Context, userId int) (domain.Room, error)

	Create(ctx context.Context, playerId, userId, scoreId int, kind domain.RoomKind) (domain.Room, error)

	DeleteById(ctx context.Context, id string) error
	DeleteByUserUnguessed(ctx context.Context, userId int) error
	DeleteByUser(ctx context.Context, userId int) error

	UpdateGuessID(ctx context.Context, id string, guessId string) error
	UpdateScore(ctx context.Context, id string, playerId, scoreId int, guessId *string) (domain.Room, error)
}

type rooms struct {
	uow *uow.UnitOfWork
}

func NewRooms(uow *uow.UnitOfWork) Rooms {
	return &rooms{uow: uow}
}

func (r *rooms) DeleteByUserUnguessed(ctx context.Context, userId int) error {
	ex := r.uow.Executor(ctx)
	_, err := ex.Exec(ctx, `
		DELETE FROM rooms WHERE user_id = $1 AND guess_id IS NULL
	`, userId)
	return err
}

func (r *rooms) DeleteByUser(ctx context.Context, userId int) error {
	ex := r.uow.Executor(ctx)
	_, err := ex.Exec(ctx, `
		DELETE FROM rooms WHERE user_id = $1 AND guess_id IS NOT NULL
	`, userId)

	return err
}

func (r *rooms) FindByUserUnguessed(ctx context.Context, userId int) (domain.Room, error) {
	ex := r.uow.Executor(ctx)
	rows, err := ex.Query(ctx, `
		SELECT * FROM rooms 
		WHERE user_id = $1 AND guess_id IS NULL
	`, userId)
	if err != nil {
		return domain.Room{}, err
	}

	return pgx.CollectOneRow(rows, rowToRoom)
}

func (r *rooms) DeleteById(ctx context.Context, id string) error {
	ex := r.uow.Executor(ctx)
	_, err := ex.Exec(ctx, "DELETE FROM rooms WHERE id = $1", id)

	return err
}

func (r *rooms) FindByUser(ctx context.Context, userId int) (domain.Room, error) {
	ex := r.uow.Executor(ctx)
	rows, err := ex.Query(ctx, `
		SELECT * FROM rooms 
		WHERE user_id = $1 AND guess_id IS NULL
		ORDER BY created_at DESC
	`, userId)
	if err != nil {
		return domain.Room{}, err
	}

	return pgx.CollectOneRow(rows, rowToRoom)
}

func (r *rooms) UpdateGuessID(ctx context.Context, id string, guessId string) error {
	ex := r.uow.Executor(ctx)
	_, err := ex.Exec(ctx, "UPDATE rooms SET guess_id = $1 WHERE id = $2", guessId, id)
	return err
}

func (r *rooms) UpdateScore(ctx context.Context, id string, playerId, scoreId int, guessId *string) (domain.Room, error) {
	ex := r.uow.Executor(ctx)
	rows, err := ex.Query(ctx, `
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

func (r *rooms) Create(ctx context.Context, playerId, userId, scoreId int, kind domain.RoomKind) (domain.Room, error) {
	ex := r.uow.Executor(ctx)
	rows, err := ex.Query(ctx, `
		INSERT INTO rooms (id, player_id, user_id, score_id, kind, closes_at) 
		VALUES (@id, @playerId, @userId, @scoreId, @kind, @closesAt) RETURNING *
	`, pgx.NamedArgs{
		"userId":   userId,
		"scoreId":  scoreId,
		"playerId": playerId,
		"kind":     kind,
		"closesAt": time.Now().Add(10 * time.Minute),
		"id":       utils.NewID(),
	})
	if err != nil {
		return domain.Room{}, err
	}

	return pgx.CollectOneRow(rows, rowToRoom)
}

func (r *rooms) FindByID(ctx context.Context, id string) (domain.Room, error) {
	ex := r.uow.Executor(ctx)
	rows, err := ex.Query(ctx, "SELECT * FROM rooms WHERE id = $1", id)
	if err != nil {
		return domain.Room{}, err
	}

	return pgx.CollectOneRow(rows, rowToRoom)
}
