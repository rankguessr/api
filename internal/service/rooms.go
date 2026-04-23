package service

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"time"

	"github.com/rankguessr/api/internal/repo"
	"github.com/rankguessr/api/internal/uow"
	"github.com/rankguessr/api/pkg/cache"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/osuapi"
	"github.com/rankguessr/api/pkg/utils"
	"github.com/redis/go-redis/v9"
)

const (
	maxGuesses          = 15
	findScoreMaxRetries = 5
	scoresLimit         = 20
)

var refillInterval = 3 * time.Minute

type Rooms interface {
	FindByUserUnguessed(ctx context.Context, userId int) (domain.Room, error)
	FindRandomScore(ctx context.Context, accessToken string) (osuapi.Score, error)
	FindByUser(ctx context.Context, userId int, accessToken string) (domain.RoomExtended, error)
	FindOrDeleteExpired(ctx context.Context, id, accessToken string, userId int) (domain.RoomExtended, error)

	UpdateGuessID(ctx context.Context, id string, guessId string) error
	SetNext(ctx context.Context, id string, userId, playerId, scoreId int) (refill domain.RefillResult, room domain.Room, err error)
	Create(ctx context.Context, playerId, userId, scoreId int, kind domain.RoomKind) (domain.RefillResult, domain.Room, error)

	DeleteById(ctx context.Context, id string) error
	DeleteByUser(ctx context.Context, userId int) error
	DeleteByUserUnguessed(ctx context.Context, userId int) error

	RefillForUser(ctx context.Context, userId int, sub uint) (domain.RefillResult, error)

	// works for now, but should be moved to a separate service
	GetScore(ctx context.Context, accessToken string, scoreId int) (osuapi.Score, error)
}

type rooms struct {
	rooms   repo.Rooms
	users   repo.Users
	players repo.Players

	guessesSvc Guess

	uow  *uow.UnitOfWork
	oapi *osuapi.Client
	rdb  *redis.Client
}

func NewRooms(
	repo repo.Rooms, users repo.Users, players repo.Players, guessesSvc Guess,
	oapi *osuapi.Client, rdb *redis.Client, uow *uow.UnitOfWork,
) Rooms {
	return &rooms{
		rooms: repo, users: users, players: players, guessesSvc: guessesSvc,
		uow: uow, oapi: oapi, rdb: rdb,
	}
}

func (s *rooms) DeleteByUserUnguessed(ctx context.Context, userId int) error {
	return s.rooms.DeleteByUserUnguessed(ctx, userId)
}

func (s *rooms) FindByUserUnguessed(ctx context.Context, userId int) (domain.Room, error) {
	return s.rooms.FindByUserUnguessed(ctx, userId)
}

func (s *rooms) DeleteByUser(ctx context.Context, userId int) error {
	return s.rooms.DeleteByUser(ctx, userId)
}

func (s *rooms) UpdateGuessID(ctx context.Context, id string, guessId string) error {
	return s.rooms.UpdateGuessID(ctx, id, guessId)
}

func (s *rooms) DeleteById(ctx context.Context, id string) error {
	return s.rooms.DeleteById(ctx, id)
}

func (r *rooms) FindRandomScore(ctx context.Context, accessToken string) (osuapi.Score, error) {
	for range findScoreMaxRetries {
		findAttempt := func() (osuapi.Score, error) {
			p, err := r.players.FindRandom(ctx)
			if err != nil {
				return osuapi.Score{}, err
			}

			scoreIdx := rand.Intn(scoresLimit)

			scores, err := r.oapi.GetUserScores(ctx, accessToken, p.OsuId, 1, scoreIdx)
			if err != nil {
				return osuapi.Score{}, err
			}

			if len(scores) == 0 {
				return osuapi.Score{}, errors.New("user has no scores")
			}

			scoreId := scores[0].ID
			// check if score exists and warm up cache
			score, err := r.GetScore(ctx, accessToken, scoreId)
			if err != nil {
				return osuapi.Score{}, err
			}

			if score.PP == 0 {
				return osuapi.Score{}, errors.New("score has 0 pp")
			}

			if score.UserId != p.OsuId {
				return osuapi.Score{}, errors.New("score user id does not match player id")
			}

			return score, nil
		}

		score, err := findAttempt()
		if err == nil {
			return score, nil
		}

		slog.WarnContext(ctx, "failed to find random score, retrying", "error", err)
	}

	return osuapi.Score{}, errors.New("failed to find random score after max retries")
}

// should be called in transaction
func (r *rooms) RefillForUser(ctx context.Context, userId int, sub uint) (domain.RefillResult, error) {
	user, err := r.users.FindForUpdate(ctx, userId)
	if err != nil {
		return domain.RefillResult{}, err
	}

	refilledAt := time.Now()
	refilled := min(maxGuesses, user.AvailableGuesses+uint((refilledAt.Sub(user.RefilledAt))/refillInterval))
	if refilled < sub {
		return domain.RefillResult{}, utils.ErrNotEnoughGuesses
	}

	newGuesses := refilled - sub
	if refilled < maxGuesses {
		rem := (refilledAt.Sub(user.RefilledAt)) % refillInterval
		refilledAt = refilledAt.Add(-rem)
	}

	err = r.users.UpdateAvailableGuesses(ctx, userId, newGuesses, refilledAt)
	if err != nil {
		return domain.RefillResult{}, err
	}

	return domain.RefillResult{
		AvailableGuesses: newGuesses,
		RefilledAt:       refilledAt,
	}, nil
}

func (r *rooms) FindByUser(ctx context.Context, userId int, accessToken string) (domain.RoomExtended, error) {
	room, err := r.rooms.FindByUser(ctx, userId)
	if err != nil {
		return domain.RoomExtended{}, err
	}

	score, err := r.GetScore(ctx, accessToken, room.ScoreID)
	if err != nil {
		err := r.DeleteById(ctx, room.ID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to delete room with invalid score", "room_id", room.ID, "score_id", room.ScoreID, "error", err)
		}

		return domain.RoomExtended{}, err
	}

	err = r.DeleteIfExpired(ctx, accessToken, userId, room, score)
	if err != nil {
		return domain.RoomExtended{}, err
	}

	return domain.RoomExtended{Room: room, Score: score}, nil
}

func (r *rooms) SetNext(ctx context.Context, id string, userId, playerId, scoreId int) (refill domain.RefillResult, room domain.Room, err error) {
	err = r.uow.WithTx(ctx, func(ctx context.Context) error {
		refill, err = r.RefillForUser(ctx, userId, 1)
		if err != nil {
			return err
		}

		room, err = r.rooms.UpdateScore(ctx, id, playerId, scoreId, nil)
		return err
	})

	return
}

func (r *rooms) DeleteIfExpired(ctx context.Context, accessToken string, userId int, room domain.Room, score osuapi.Score) error {
	if room.ClosesAt.Before(time.Now()) && room.GuessID == nil {
		player, err := r.oapi.GetUser(ctx, accessToken, score.User.ID)
		if err != nil {
			return err
		}

		_, _, err = r.guessesSvc.CreateAndUpdateUserElo(
			ctx, userId, domain.GuessCreate{
				PlayerID:     player.ID,
				Guess:        0,
				ScoreID:      room.ScoreID,
				BeatmapID:    score.BeatmapID,
				BeatmapSetID: score.Beatmap.BeatmapSetId,
				Kind:         room.Kind,
				ActualRank:   player.Statistics.GlobalRank,
			},
		)
		if err != nil {
			return err
		}

		err = r.rooms.DeleteById(ctx, room.ID)
		if err != nil {
			return err
		}

		return utils.ErrRoomClosed
	}

	return nil
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

func (r *rooms) FindOrDeleteExpired(ctx context.Context, id, accessToken string, userId int) (domain.RoomExtended, error) {
	room, err := r.rooms.FindByID(ctx, id)
	if err != nil {
		return domain.RoomExtended{}, err
	}

	score, err := r.GetScore(ctx, accessToken, room.ScoreID)
	if err != nil {
		return domain.RoomExtended{}, err
	}

	err = r.DeleteIfExpired(ctx, accessToken, userId, room, score)
	if err != nil {
		return domain.RoomExtended{}, err
	}

	return domain.RoomExtended{Room: room, Score: score}, nil
}

func (s *rooms) Create(ctx context.Context, playerId, userId, scoreId int, kind domain.RoomKind) (refill domain.RefillResult, room domain.Room, err error) {
	err = s.uow.WithTx(ctx, func(ctx context.Context) error {
		refill, err = s.RefillForUser(ctx, userId, 1)
		if err != nil {
			return err
		}

		room, err = s.rooms.Create(ctx, playerId, userId, scoreId, kind)
		return err
	})
	if err != nil {
		return domain.RefillResult{}, domain.Room{}, err
	}

	return refill, room, nil
}
