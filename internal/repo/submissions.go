package repo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/rankguessr/api/internal/uow"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/utils"
)

var (
	rowToSubmission         = pgx.RowToStructByName[domain.Submission]
	rowToSubmissionExtended = pgx.RowToStructByName[domain.SubmissionExtended]
)

type Submissions interface {
	Create(ctx context.Context, input domain.SubmissionCreate) (domain.Submission, error)
	Delete(ctx context.Context, id string) error
	SetAccepted(ctx context.Context, id string) error

	FindRandom(ctx context.Context, userId int) (domain.Submission, error)
	FindByUser(ctx context.Context, userId int, accepted bool) ([]domain.Submission, error)
	Find(ctx context.Context, accepted bool, limit, offset int) ([]domain.SubmissionExtended, error)
}

type submissions struct {
	uow *uow.UnitOfWork
}

func NewSubmissions(uow *uow.UnitOfWork) Submissions {
	return &submissions{uow: uow}
}

func (s *submissions) Create(ctx context.Context, input domain.SubmissionCreate) (domain.Submission, error) {
	ex := s.uow.Executor(ctx)
	rows, err := ex.Query(ctx, `
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
	ex := s.uow.Executor(ctx)
	_, err := ex.Exec(ctx, "DELETE FROM submissions WHERE id = $1", id)
	return err
}

func (s *submissions) FindByUser(ctx context.Context, userId int, accepted bool) ([]domain.Submission, error) {
	ex := s.uow.Executor(ctx)
	rows, err := ex.Query(ctx, "SELECT * FROM submissions WHERE user_id = $1 AND is_accepted = $2", userId, accepted)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, rowToSubmission)
}

func (s *submissions) FindRandom(ctx context.Context, userId int) (domain.Submission, error) {
	ex := s.uow.Executor(ctx)
	rows, err := ex.Query(ctx, `
		SELECT * FROM submissions s 
		WHERE NOT EXISTS (
			SELECT 1 FROM guesses g
			WHERE g.kind = 'v2sub'
			AND g.score_id = s.score_id AND g.user_id = $1
		) AND s.user_id != $1
		ORDER BY RANDOM() LIMIT 1
	`, userId)
	if err != nil {
		return domain.Submission{}, err
	}

	return pgx.CollectOneRow(rows, rowToSubmission)
}

func (s *submissions) Find(ctx context.Context, accepted bool, limit, offset int) ([]domain.SubmissionExtended, error) {
	ex := s.uow.Executor(ctx)
	rows, err := ex.Query(ctx, `
		SELECT 
			s.*, to_json(u) AS user 
		FROM submissions s  
		JOIN users u ON s.user_id = u.osu_id
		WHERE is_accepted = $1 
		LIMIT $2 OFFSET $3
	`, accepted, limit, offset)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, rowToSubmissionExtended)
}

func (s *submissions) SetAccepted(ctx context.Context, id string) error {
	ex := s.uow.Executor(ctx)
	_, err := ex.Exec(ctx, "UPDATE submissions SET is_accepted = TRUE, updated_at = NOW() WHERE id = $1", id)
	return err
}
