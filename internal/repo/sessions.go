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
	Find(ctx context.Context, id, key string) (domain.Session, error)
	FindWithUser(ctx context.Context, id, key string) (domain.SessionExtended, error)

	DeleteByUser(ctx context.Context, osuId int) error
	Create(ctx context.Context, osuId int, accessToken, refreshToken string, expiresAt time.Time, key string) (domain.Session, error)
	UpdateTokens(ctx context.Context, id, accessToken, refreshToken string, expiresIn time.Time, key string) error
}

type sessions struct {
	pool *pgxpool.Pool
}

func NewSessions(pool *pgxpool.Pool) Sessions {
	return &sessions{pool: pool}
}

func (s *sessions) FindWithUser(ctx context.Context, id, key string) (domain.SessionExtended, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT 
			s.id, 
			pgp_sym_decrypt(s.access_token, TEXT($2)) AS access_token,
			pgp_sym_decrypt(s.refresh_token, TEXT($2)) AS refresh_token, 
			s.expires_at, s.created_at, to_json(u) AS user
		FROM sessions s
		JOIN users u ON s.user_id = u.osu_id
		WHERE s.id = $1
	`, id, key)
	if err != nil {
		return domain.SessionExtended{}, err
	}

	return pgx.CollectExactlyOneRow(rows, rowToSessionExtended)
}

func (s *sessions) Create(ctx context.Context, osuId int, accessToken string, refreshToken string, expiresAt time.Time, key string) (domain.Session, error) {
	rows, err := s.pool.Query(ctx, `
		INSERT INTO sessions (id, user_id, access_token, refresh_token, expires_at)
		VALUES 
			(@id, @userId, 
			pgp_sym_encrypt(@accessToken, TEXT(@encKey)), 
			pgp_sym_encrypt(@refreshToken, TEXT(@encKey)), 
			@expiresAt)
		RETURNING id, user_id, expires_at, created_at,
			pgp_sym_decrypt(access_token, TEXT(@encKey)) AS access_token,
			pgp_sym_decrypt(refresh_token, TEXT(@encKey)) AS refresh_token
	`, pgx.NamedArgs{
		"id":           utils.NewID(),
		"userId":       osuId,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"expiresAt":    expiresAt,
		"encKey":       key,
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

func (s *sessions) Find(ctx context.Context, id, key string) (domain.Session, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, expires_at, created_at,
			TEXT(pgp_sym_decrypt(s.access_token, TEXT($2))) AS access_token,
			TEXT(pgp_sym_decrypt(s.refresh_token, TEXT($2))) AS refresh_token
		FROM sessions WHERE id = $1
	`, id, key)
	if err != nil {
		return domain.Session{}, err
	}

	return pgx.CollectExactlyOneRow(rows, rowToSession)
}

func (s *sessions) UpdateTokens(ctx context.Context, id, accessToken, refreshToken string, expiresIn time.Time, key string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE sessions 
		SET access_token = pgp_sym_encrypt(@accessToken, TEXT(@encKey)),
			refresh_token = pgp_sym_encrypt(@refreshToken, TEXT(@encKey)),
			expires_at = @expiresAt
		WHERE id = @id
	`, pgx.NamedArgs{
		"id":           id,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"expiresAt":    expiresIn,
		"encKey":       key,
	})
	return err
}
