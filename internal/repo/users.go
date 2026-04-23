package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rankguessr/api/internal/uow"
	"github.com/rankguessr/api/pkg/domain"
)

var (
	rowToUser         = pgx.RowToStructByName[domain.User]
	rowToUserExtended = pgx.RowToStructByName[domain.UserExtended]
)

type Users interface {
	UpdateElo(ctx context.Context, userId, elo int) (int, error)
	Upsert(ctx context.Context, osuId int, username, avatarURL, countryCode string) error
	UpdateAvailableGuesses(ctx context.Context, userId int, guesses uint, refilledAt time.Time) error

	FindTop(ctx context.Context, limit, offset int) ([]domain.UserExtended, error)
	FindByOsuID(ctx context.Context, osuId int) (domain.User, error)
	FindForUpdate(ctx context.Context, osuId int) (domain.User, error)
}

type users struct {
	uow *uow.UnitOfWork
}

func NewUsers(uow *uow.UnitOfWork) Users {
	return &users{uow: uow}
}

func (u *users) UpdateAvailableGuesses(ctx context.Context, userId int, guesses uint, refilledAt time.Time) error {
	ex := u.uow.Executor(ctx)
	_, err := ex.Exec(ctx, `
		UPDATE users 
		SET available_guesses = $1, refilled_at = $2
		WHERE osu_id = $3
	`, guesses, refilledAt, userId)
	return err
}

func (u *users) FindForUpdate(ctx context.Context, osuId int) (domain.User, error) {
	ex := u.uow.Executor(ctx)
	rows, err := ex.Query(ctx, `
		SELECT * FROM users 
		WHERE osu_id = $1
		FOR UPDATE
	`, osuId)
	if err != nil {
		return domain.User{}, err
	}

	return pgx.CollectOneRow(rows, rowToUser)
}

// adds the given elo to the user's current elo
func (u *users) UpdateElo(ctx context.Context, userId int, elo int) (int, error) {
	ex := u.uow.Executor(ctx)
	rows, err := ex.Query(ctx, "UPDATE users SET elo = GREATEST(0, elo + $1) WHERE osu_id = $2 RETURNING elo", elo, userId)
	if err != nil {
		return 0, err
	}

	return pgx.CollectOneRow(rows, pgx.RowTo[int])
}

func (u *users) FindByOsuID(ctx context.Context, osuId int) (domain.User, error) {
	ex := u.uow.Executor(ctx)
	rows, err := ex.Query(ctx, "SELECT * FROM users WHERE osu_id = $1", osuId)
	if err != nil {
		return domain.User{}, err
	}

	return pgx.CollectOneRow(rows, rowToUser)
}

func (u *users) FindTop(ctx context.Context, limit, offset int) ([]domain.UserExtended, error) {
	ex := u.uow.Executor(ctx)
	rows, err := ex.Query(ctx, `
		SELECT u.*, 
			COUNT(g) AS total_guesses,
			ROW_NUMBER() OVER (ORDER BY u.elo DESC) as rank
		FROM users u 
		JOIN guesses g ON u.osu_id = g.user_id
		WHERE g.kind = 'v2'
		GROUP BY u.osu_id
		ORDER BY u.elo DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, rowToUserExtended)
}

func (u *users) Upsert(ctx context.Context, osuId int, username, avatarURL, countryCode string) error {
	ex := u.uow.Executor(ctx)
	_, err := ex.Exec(ctx, `
		INSERT INTO users 
			(osu_id, username, avatar_url, country_code)
		VALUES
			(@osu_id, @username, @avatar_url, @country_code)
		ON CONFLICT (osu_id) DO UPDATE
		SET username = EXCLUDED.username,
			avatar_url = EXCLUDED.avatar_url,
			country_code = EXCLUDED.country_code
	`, pgx.NamedArgs{
		"osu_id":       osuId,
		"username":     username,
		"avatar_url":   avatarURL,
		"country_code": countryCode,
	})

	return err
}
