package repo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rankguessr/api/pkg/domain"
)

var rowToUser = pgx.RowToStructByName[domain.User]

type Users interface {
	Upsert(ctx context.Context, osuId int, username, avatarURL, countryCode string) error

	FindByOsuID(ctx context.Context, osuId int) (domain.User, error)
}

type users struct {
	pool *pgxpool.Pool
}

func NewUsers(pool *pgxpool.Pool) Users {
	return &users{pool: pool}
}

const findByOsuIDQuery = `
	SELECT *
	FROM users
	WHERE osu_id = $1
`

func (u *users) FindByOsuID(ctx context.Context, osuId int) (domain.User, error) {
	rows, err := u.pool.Query(ctx, findByOsuIDQuery, osuId)
	if err != nil {
		return domain.User{}, err
	}

	return pgx.CollectOneRow(rows, rowToUser)
}

const upsertUserQuery = `
	INSERT INTO users 
		(osu_id, username, avatar_url, country_code)
	VALUES
		(@osu_id, @username, @avatar_url, @country_code)
	ON CONFLICT (osu_id) DO UPDATE
	SET username = EXCLUDED.username,
		avatar_url = EXCLUDED.avatar_url,
		country_code = EXCLUDED.country_code
`

func (u *users) Upsert(ctx context.Context, osuId int, username, avatarURL, countryCode string) error {
	_, err := u.pool.Exec(ctx, upsertUserQuery, pgx.NamedArgs{
		"osu_id":       osuId,
		"username":     username,
		"avatar_url":   avatarURL,
		"country_code": countryCode,
	})

	return err
}
