package service

import (
	"context"

	"github.com/rankguessr/api/internal/repo"
	"github.com/rankguessr/api/pkg/domain"
)

type Rooms interface {
	FindByID(ctx context.Context, id string) (domain.Room, error)
	Close(ctx context.Context, id string) error
	Create(ctx context.Context, playerId, userId, scoreId int) (domain.Room, error)
}

type rooms struct {
	repo repo.Rooms
}

func NewRooms(repo repo.Rooms) Rooms {
	return &rooms{repo: repo}
}

func (s *rooms) Close(ctx context.Context, id string) error {
	return s.repo.SetClosed(ctx, id, true)
}

func (s *rooms) FindByID(ctx context.Context, id string) (domain.Room, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *rooms) Create(ctx context.Context, playerId int, userId int, scoreId int) (domain.Room, error) {
	return s.repo.Create(ctx, playerId, userId, scoreId)
}
