package service

import (
	"context"

	"github.com/rankguessr/api/internal/repo"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/osuapi"
)

type Submissions interface {
	Create(ctx context.Context, input domain.SubmissionCreate) (domain.Submission, error)
	Delete(ctx context.Context, id string) error
	SetAccepted(ctx context.Context, id string) error

	FindRandom(ctx context.Context, userId int) (domain.Submission, error)
	FindRandomWithScore(ctx context.Context, userId int, accessToken string) (osuapi.Score, error)
	FindByUser(ctx context.Context, userId int, accepted bool) ([]domain.Submission, error)
	Find(ctx context.Context, accepted bool, limit, page int) ([]domain.SubmissionExtended, error)
}

type submissions struct {
	repo   repo.Submissions
	client *osuapi.Client
}

func NewSubmissions(repo repo.Submissions, client *osuapi.Client) Submissions {
	return &submissions{repo: repo, client: client}
}

func (s *submissions) Create(ctx context.Context, input domain.SubmissionCreate) (domain.Submission, error) {
	return s.repo.Create(ctx, input)
}

func (s *submissions) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *submissions) FindByUser(ctx context.Context, userId int, accepted bool) ([]domain.Submission, error) {
	return s.repo.FindByUser(ctx, userId, accepted)
}

func (s *submissions) FindRandom(ctx context.Context, userId int) (domain.Submission, error) {
	return s.repo.FindRandom(ctx, userId)
}

func (s *submissions) FindRandomWithScore(ctx context.Context, userId int, accessToken string) (osuapi.Score, error) {
	sub, err := s.repo.FindRandom(ctx, userId)
	if err != nil {
		return osuapi.Score{}, err
	}

	return s.client.GetScore(ctx, accessToken, sub.ScoreID)
}

func (s *submissions) Find(ctx context.Context, accepted bool, limit, page int) ([]domain.SubmissionExtended, error) {
	return s.repo.Find(ctx, accepted, limit, limit*(max(page, 1)-1))
}

func (s *submissions) SetAccepted(ctx context.Context, id string) error {
	return s.repo.SetAccepted(ctx, id)
}
