package service

import (
	"context"

	"github.com/rankguessr/api/internal/repo"
	"github.com/rankguessr/api/pkg/domain"
)

type Submissions interface {
	Create(ctx context.Context, userId, playerId, scoreId int) (domain.Submission, error)
	Delete(ctx context.Context, id string) error
	SetAccepted(ctx context.Context, id string) error

	FindRandom(ctx context.Context, userId int) (domain.Submission, error)
	FindByUser(ctx context.Context, userId int) ([]domain.Submission, error)
	FindUnaccepted(ctx context.Context) ([]domain.Submission, error)
}

type submissions struct {
	repo repo.Submissions
}

func NewSubmissions(repo repo.Submissions) Submissions {
	return &submissions{repo: repo}
}

func (s *submissions) Create(ctx context.Context, userId int, playerId int, scoreId int) (domain.Submission, error) {
	return s.repo.Create(ctx, userId, playerId, scoreId)
}

func (s *submissions) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *submissions) FindByUser(ctx context.Context, userId int) ([]domain.Submission, error) {
	return s.repo.FindByUser(ctx, userId)
}

func (s *submissions) FindRandom(ctx context.Context, userId int) (domain.Submission, error) {
	return s.repo.FindRandom(ctx, userId)
}

func (s *submissions) FindUnaccepted(ctx context.Context) ([]domain.Submission, error) {
	return s.repo.FindUnaccepted(ctx)
}

func (s *submissions) SetAccepted(ctx context.Context, id string) error {
	return s.repo.SetAccepted(ctx, id)
}
