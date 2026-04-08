package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/utils"
)

var rowToSession = pgx.RowToStructByName[domain.Session]
var rowToSessionExtended = pgx.RowToStructByName[domain.SessionExtended]

type Sessions interface {
	Find(ctx context.Context, id string) (domain.Session, error)
	FindWithUser(ctx context.Context, id string) (domain.SessionExtended, error)

	DeleteByUser(ctx context.Context, osuId int) error
	Create(ctx context.Context, osuId int, accessToken, refreshToken string, expiresAt time.Time) (domain.Session, error)
	UpdateTokens(ctx context.Context, id, accessToken, refreshToken string, expiresIn time.Time) error
}

type sessions struct {
	pool *pgxpool.Pool
}

func NewSessions(pool *pgxpool.Pool) Sessions {
	return &sessions{pool: pool}
}

func (s *sessions) FindWithUser(ctx context.Context, id string) (domain.SessionExtended, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT 
			s.id, s.access_token, s.refresh_token, s.expires_at, s.created_at,
			to_json(u) AS user
		FROM sessions s
		JOIN users u ON s.user_id = u.osu_id
		WHERE s.id = $1
	`, id)
	if err != nil {
		return domain.SessionExtended{}, err
	}

	return pgx.CollectExactlyOneRow(rows, rowToSessionExtended)
}

func (s *sessions) Create(ctx context.Context, osuId int, accessToken string, refreshToken string, expiresAt time.Time) (domain.Session, error) {
	rows, err := s.pool.Query(ctx, `
		INSERT INTO sessions (id, user_id, access_token, refresh_token, expires_at)
		VALUES (@id, @userId, @accessToken, @refreshToken, @expiresAt)
		RETURNING *
	`, pgx.NamedArgs{
		"id":           utils.NewID(),
		"userId":       osuId,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"expiresAt":    expiresAt,
	})
	if err != nil {
		return domain.Session{}, err
	}

	return pgx.CollectExactlyOneRow(rows, rowToSession)
}

func (s *sessions) DeleteByUser(ctx context.Context, osuId int) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM sessions WHERE user_id = $1", osuId)
	return err
}

func (s *sessions) Find(ctx context.Context, id string) (domain.Session, error) {
	rows, err := s.pool.Query(ctx, "SELECT * FROM sessions WHERE id = $1", id)
	if err != nil {
		return domain.Session{}, err
	}

	return pgx.CollectExactlyOneRow(rows, rowToSession)
}

func (s *sessions) UpdateTokens(ctx context.Context, id, accessToken, refreshToken string, expiresIn time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE sessions 
		SET access_token = $1, 
			refresh_token = $2, 
			expires_at = $4
		WHERE id = $3
	`, accessToken, refreshToken, id, expiresIn)
	return err
}
