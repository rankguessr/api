package repo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/utils"
)

var rowToSubmission = pgx.RowToStructByName[domain.Submission]

type Submissions interface {
	Create(ctx context.Context, input domain.SubmissionCreate) (domain.Submission, error)
	Delete(ctx context.Context, id string) error
	SetAccepted(ctx context.Context, id string) error

	FindRandom(ctx context.Context, userId int) (domain.Submission, error)
	FindByUser(ctx context.Context, userId int) ([]domain.Submission, error)
	FindUnaccepted(ctx context.Context) ([]domain.Submission, error)
}

type submissions struct {
	pool *pgxpool.Pool
}

func NewSubmissions(pool *pgxpool.Pool) Submissions {
	return &submissions{pool: pool}
}

func (s *submissions) Create(ctx context.Context, input domain.SubmissionCreate) (domain.Submission, error) {
	rows, err := s.pool.Query(ctx, `
		INSERT INTO submissions (id, user_id, player_id, score_id, comment, beatmap_id, beatmapset_id) 
		VALUES (@id, @userId, @playerId, @scoreId, @comment, @beatmapId, @beatmapsetId) RETURNING *
	`, pgx.NamedArgs{
		"id":           utils.NewID(),
		"userId":       input.UserID,
		"playerId":     input.PlayerID,
		"scoreId":      input.ScoreID,
		"comment":      input.Comment,
		"beatmapId":    input.BeatmapID,
		"beatmapsetId": input.BeatmapsetID,
	})
	if err != nil {
		return domain.Submission{}, err
	}

	return pgx.CollectOneRow(rows, rowToSubmission)
}

func (s *submissions) Delete(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM submissions WHERE id = $1", id)
	return err
}

func (s *submissions) FindByUser(ctx context.Context, userId int) ([]domain.Submission, error) {
	rows, err := s.pool.Query(ctx, "SELECT * FROM submissions WHERE user_id = $1", userId)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, rowToSubmission)
}

func (s *submissions) FindRandom(ctx context.Context, userId int) (domain.Submission, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT * FROM submissions s 
		WHERE NOT EXISTS (
			SELECT 1 FROM guesses g
			WHERE g.is_submission 
			AND g.score_id = s.score_id AND g.user_id = $1
		) AND s.user_id != $1
		ORDER BY RANDOM() LIMIT 1
	`, userId)
	if err != nil {
		return domain.Submission{}, err
	}

	return pgx.CollectOneRow(rows, rowToSubmission)
}

func (s *submissions) FindUnaccepted(ctx context.Context) ([]domain.Submission, error) {
	rows, err := s.pool.Query(ctx, "SELECT * FROM submissions WHERE NOT is_accepted")
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, rowToSubmission)
}

func (s *submissions) SetAccepted(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, "UPDATE submissions SET is_accepted = TRUE, updated_at = NOW() WHERE id = $1", id)
	return err
}
