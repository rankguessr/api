package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/rankguessr/api/internal/repo"
	"github.com/rankguessr/api/pkg/cache"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/osuapi"
	"github.com/redis/go-redis/v9"
)

type Rooms interface {
	FindByID(ctx context.Context, id string) (domain.Room, error)
	FindByUser(ctx context.Context, userId int) (domain.Room, error)
	FindByUserUnguessed(ctx context.Context, userId int) (domain.Room, error)

	UpdateGuessID(ctx context.Context, id string, guessId string) error
	UpdateScore(ctx context.Context, id string, playerId, scoreId int) (domain.Room, error)
	Create(ctx context.Context, playerId, userId, scoreId int) (domain.Room, error)

	DeleteById(ctx context.Context, id string) error
	DeleteByUser(ctx context.Context, userId int) error
	DeleteByUserUnguessed(ctx context.Context, userId int) error

	RefillForUser(ctx context.Context, userId int, sub uint) (domain.RefillResult, error)

	// works for now, but should be moved to a separate service
	GetScore(ctx context.Context, accessToken string, scoreId int) (osuapi.Score, error)
}

type rooms struct {
	repo repo.Rooms
	oapi *osuapi.Client
	rdb  *redis.Client
}

func NewRooms(repo repo.Rooms, oapi *osuapi.Client, rdb *redis.Client) Rooms {
	return &rooms{repo: repo, oapi: oapi, rdb: rdb}
}

func (s *rooms) RefillForUser(ctx context.Context, userId int, sub uint) (domain.RefillResult, error) {
	return s.repo.RefillForUser(ctx, userId, sub)
}

func (s *rooms) GetScore(ctx context.Context, accessToken string, scoreId int) (osuapi.Score, error) {
	score, err := cache.GetScore(s.rdb, ctx, scoreId)
	if err == nil {
		return score, nil
	}

	score, err = s.oapi.GetScore(ctx, accessToken, scoreId)
	if err != nil {
		return osuapi.Score{}, err
	}

	err = cache.SetScore(s.rdb, ctx, score)
	if err != nil {
		slog.ErrorContext(ctx, "failed to set score cache")
	}

	return score, nil
}

func (s *rooms) DeleteByUserUnguessed(ctx context.Context, userId int) error {
	return s.repo.DeleteByUserUnguessed(ctx, userId)
}

func (s *rooms) FindByUserUnguessed(ctx context.Context, userId int) (domain.Room, error) {
	return s.repo.FindByUserUnguessed(ctx, userId)
}

func (s *rooms) DeleteByUser(ctx context.Context, userId int) error {
	return s.repo.DeleteByUser(ctx, userId)
}

func (s *rooms) UpdateGuessID(ctx context.Context, id string, guessId string) error {
	return s.repo.UpdateGuessID(ctx, id, guessId)
}

func (s *rooms) DeleteById(ctx context.Context, id string) error {
	return s.repo.DeleteById(ctx, id)
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

func (s *rooms) Create(ctx context.Context, playerId, userId, scoreId int) (domain.Room, error) {
	_, err := s.FindByUser(ctx, userId)
	if err == nil {
		return domain.Room{}, errors.New("user already has a room")
	}

	return s.repo.Create(ctx, playerId, userId, scoreId)
}
