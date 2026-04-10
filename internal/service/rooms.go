package service

import (
	"context"
	"errors"

	"github.com/rankguessr/api/internal/repo"
	"github.com/rankguessr/api/pkg/domain"
)

type Rooms interface {
	FindByID(ctx context.Context, id string) (domain.Room, error)
	FindByUser(ctx context.Context, userId int) (domain.Room, error)

	UpdateGuessID(ctx context.Context, id string, guessId string) error
	UpdateScore(ctx context.Context, id string, playerId, scoreId int) (domain.Room, error)
	Create(ctx context.Context, playerId, userId, scoreId int) (domain.Room, error)
}

type rooms struct {
	repo repo.Rooms
}

func NewRooms(repo repo.Rooms) Rooms {
	return &rooms{repo: repo}
}

func (s *rooms) UpdateGuessID(ctx context.Context, id string, guessId string) error {
	return s.repo.UpdateGuessID(ctx, id, guessId)
}

func (s *rooms) FindByUser(ctx context.Context, userId int) (domain.Room, error) {
	return s.repo.FindByUser(ctx, userId)
}

func (s *rooms) UpdateScore(ctx context.Context, id string, playerId, scoreId int) (domain.Room, error) {
	return s.repo.UpdateScore(ctx, id, playerId, scoreId, nil)
}

func (s *rooms) FindByID(ctx context.Context, id string) (domain.Room, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *rooms) Create(ctx context.Context, playerId int, userId int, scoreId int) (domain.Room, error) {
	_, err := s.FindByUser(ctx, userId)
	if err == nil {
		return domain.Room{}, errors.New("user already has a room")
	}

	return s.repo.Create(ctx, playerId, userId, scoreId)
}
