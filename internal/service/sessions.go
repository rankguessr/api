package service

import (
	"context"
	"time"

	"github.com/rankguessr/api/internal/repo"
	"github.com/rankguessr/api/pkg/domain"
)

type Sessions interface {
	Find(ctx context.Context, id string) (domain.Session, error)
	FindWithUser(ctx context.Context, id string) (domain.SessionExtended, error)

	DeleteByUser(ctx context.Context, osuId int) error
	UpdateTokens(ctx context.Context, id, accessToken, refreshToken string, expiresAt time.Time) error
	Create(ctx context.Context, osuId int, accessToken, refreshToken string, expiresAt time.Time) (domain.Session, error)
}

type sessions struct {
	repo repo.Sessions
}

func NewSessions(repo repo.Sessions) Sessions {
	return &sessions{repo: repo}
}

func (s *sessions) Create(ctx context.Context, osuId int, accessToken string, refreshToken string, expiresAt time.Time) (domain.Session, error) {
	err := s.repo.DeleteByUser(ctx, osuId)
	if err != nil {
		return domain.Session{}, err
	}

	return s.repo.Create(ctx, osuId, accessToken, refreshToken, expiresAt)
}

func (s *sessions) DeleteByUser(ctx context.Context, osuId int) error {
	return s.repo.DeleteByUser(ctx, osuId)
}

func (s *sessions) Find(ctx context.Context, id string) (domain.Session, error) {
	return s.repo.Find(ctx, id)
}

func (s *sessions) FindWithUser(ctx context.Context, id string) (domain.SessionExtended, error) {
	return s.repo.FindWithUser(ctx, id)
}

func (s *sessions) UpdateTokens(ctx context.Context, id, accessToken, refreshToken string, expiresAt time.Time) error {
	return s.repo.UpdateTokens(ctx, id, accessToken, refreshToken, expiresAt)
}
