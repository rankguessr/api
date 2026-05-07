package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rankguessr/api/internal/config"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/osuapi"
	"github.com/rankguessr/api/pkg/ranking"
	"github.com/rankguessr/api/pkg/utils"
)

const RoomStartMaxRetries = 5
const ScoresLimit = 20

func RoomStart(player service.Players, rooms service.Rooms, subs service.Submissions) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return echo.ErrUnauthorized.Wrap(err)
		}

		var req struct {
			Kind domain.RoomKind `json:"kind"`
		}

		if err := c.Bind(&req); err != nil {
			return echo.ErrBadRequest.Wrap(err)
		}

		_, err = rooms.FindByUserUnguessed(ctx, session.User.OsuID)
		if err == nil {
			return echo.NewHTTPError(http.StatusBadRequest, "user is already in room")
		}

		err = rooms.DeleteByUser(ctx, session.User.OsuID)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		var score osuapi.Score
		if req.Kind == domain.RoomKindRankedV2 {
			score, err = rooms.FindRandomScore(ctx, session.AccessToken)
		} else {
			score, _, err = subs.FindRandomWithScore(ctx, session.User.OsuID, session.AccessToken)
		}

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "failed to find a score, try again later").Wrap(err)
		}

		refill, room, err := rooms.Create(ctx, score.User.ID, session.User.OsuID, score.ID, req.Kind)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "failed to create a room, try again later").Wrap(err)
		}

		return c.JSON(http.StatusOK, utils.Map{
			"refill":  refill,
			"room_id": room.ID,
		})
	}
}

func RoomGetNext(rooms service.Rooms, players service.Players, subs service.Submissions) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return err
		}

		roomId := c.Param("id")

		room, err := rooms.FindOrDeleteExpired(ctx, roomId, session.AccessToken, session.User.OsuID)
		if err != nil {
			return echo.ErrNotFound.Wrap(err)
		}

		if room.UserID != session.User.OsuID {
			return echo.ErrUnauthorized
		}

		comment := ""

		var score osuapi.Score
		if room.Kind == domain.RoomKindRankedV2 {
			score, err = rooms.FindRandomScore(ctx, session.AccessToken)
		} else {
			score, comment, err = subs.FindRandomWithScore(ctx, session.User.OsuID, session.AccessToken)
		}
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "failed to find a score, try again later").Wrap(err)
		}

		refill, newRoom, err := rooms.SetNext(ctx, roomId, session.User.OsuID, score.User.ID, score.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to set next score").Wrap(err)
		}

		return c.JSON(200, utils.Map{
			"score": utils.Map{
				"pp":         score.PP,
				"mods":       score.ModsAcronyms(),
				"accuracy":   score.Accuracy,
				"beatmapset": score.BeatmapSet,
				"beatmap":    score.Beatmap,
				"statistics": score.Statistics,
			},
			"refill":    refill,
			"comment":   comment,
			"closes_at": newRoom.ClosesAt,
		})
	}
}

func RoomPrepareReplay(rooms service.Rooms, client *osuapi.Client, replays service.Replays) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return err
		}

		roomId := c.Param("id")
		room, err := rooms.FindOrDeleteExpired(ctx, roomId, session.AccessToken, session.User.OsuID)
		if err != nil {
			return echo.ErrNotFound.Wrap(err)
		}

		if room.UserID != session.User.OsuID {
			return echo.ErrUnauthorized
		}

		if room.ReplayURL != "" {
			return c.JSON(200, utils.Map{
				"url": room.ReplayURL,
			})
		}

		replay, err := client.DownloadReplay(ctx, session.AccessToken, room.ScoreID)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		ttl := time.Until(room.ClosesAt)
		filename := fmt.Sprintf("%s.osr", room.ID)
		presigned, err := replays.AnonymizeAndPresign(ctx, replay, filename, ttl)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		err = rooms.UpdateReplayURL(ctx, room.ID, presigned)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		return c.JSON(200, utils.Map{
			"url": presigned,
		})
	}
}

func RoomDownloadReplay(rooms service.Rooms, replays service.Replays, client *osuapi.Client) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return err
		}

		filename := c.Param("filename")

		roomId := strings.TrimSuffix(filename, ".osr")
		room, err := rooms.FindOrDeleteExpired(ctx, roomId, session.AccessToken, session.User.OsuID)
		if err != nil {
			return echo.ErrNotFound.Wrap(err)
		}

		if room.ReplayURL != "" {
			return c.Redirect(http.StatusFound, room.ReplayURL)
		}

		replay, err := client.DownloadReplay(ctx, session.AccessToken, room.ScoreID)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		ttl := time.Until(room.ClosesAt)
		presigned, err := replays.AnonymizeAndPresign(ctx, replay, filename, ttl)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		err = rooms.UpdateReplayURL(ctx, roomId, presigned)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		return c.Redirect(http.StatusFound, presigned)
	}
}

func RoomGetScore(rooms service.Rooms, guesses service.Guess, subs service.Submissions) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return err
		}

		roomId := c.Param("id")
		room, err := rooms.FindOrDeleteExpired(ctx, roomId, session.AccessToken, session.User.OsuID)
		if err != nil {
			return echo.ErrNotFound.Wrap(err)
		}

		if room.UserID != session.User.OsuID {
			return echo.ErrUnauthorized
		}

		var guess *domain.Guess
		if room.GuessID != nil {
			g, err := guesses.FindById(ctx, string(*room.GuessID))
			if err != nil {
				return err
			}
			guess = &g
		}

		var user *osuapi.User
		if guess != nil {
			user = &room.Score.User
		}

		comment := ""
		if room.Kind == domain.RoomKindSubmissionV2 {
			sub, err := subs.FindByScoreID(ctx, room.ScoreID)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "failed to find a score, try again later").Wrap(err)
			}
			comment = sub.Comment
		}

		return c.JSON(200, utils.Map{
			"score": utils.Map{
				"pp":         room.Score.PP,
				"mods":       room.Score.ModsAcronyms(),
				"accuracy":   room.Score.Accuracy,
				"beatmapset": room.Score.BeatmapSet,
				"beatmap":    room.Score.Beatmap,
				"statistics": room.Score.Statistics,
				"user":       user,
			},
			"comment":   comment,
			"kind":      room.Kind,
			"closes_at": room.ClosesAt,
			"guess":     guess,
		})
	}
}

type submitRequest struct {
	Guess int    `json:"guess"`
	Token string `json:"token"`
}

func RoomSubmitGuess(rooms service.Rooms, guesses service.Guess, client *osuapi.Client, cfg *config.Config) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return echo.ErrUnauthorized.Wrap(err)
		}

		var req submitRequest
		if err := c.Bind(&req); err != nil {
			return echo.ErrBadRequest.Wrap(err)
		}

		err = utils.ValidateTurnstile(req.Token, cfg.TurnstileSecret)
		if err != nil {
			var tuErr *utils.TurnstileError
			if errors.As(err, &tuErr) {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("turnstile validation error: %s", tuErr.Error()))
			}

			return echo.ErrInternalServerError.Wrap(err)
		}

		roomId := c.Param("id")
		room, err := rooms.FindOrDeleteExpired(ctx, roomId, session.AccessToken, session.User.OsuID)
		if err != nil {
			return echo.ErrNotFound.Wrap(err)
		}

		if req.Guess > 3000000 || req.Guess < 1 {
			return echo.NewHTTPError(http.StatusBadRequest, "guess must be between 1 and 3 million")
		}

		if room.UserID != session.User.OsuID {
			return echo.ErrUnauthorized
		}

		if room.GuessID != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "room is already closed")
		}

		player, err := client.GetUser(ctx, session.AccessToken, room.Score.User.ID)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		elo, guess, err := guesses.CreateAndUpdateUserElo(
			ctx, session.User.OsuID, domain.GuessCreate{
				PlayerID:     player.ID,
				Guess:        req.Guess,
				ScoreID:      room.ScoreID,
				BeatmapID:    room.Score.BeatmapID,
				BeatmapSetID: room.Score.Beatmap.BeatmapSetId,
				ActualRank:   player.Statistics.GlobalRank,
				Kind:         room.Kind,
			},
		)

		// this should never happen with new pool
		// TODO: check player rank against ranges before creating a room, it's cached anyway
		if errors.Is(err, ranking.ErrRangeNotFound) {
			err := rooms.DeleteById(ctx, roomId)
			if err != nil {
				return echo.ErrInternalServerError.Wrap(err)
			}

			return echo.NewHTTPError(http.StatusInternalServerError, "actual rank is out of range, deleted the room, please try again")
		}

		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		err = rooms.UpdateGuessID(ctx, room.ID, guess.ID)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		return c.JSON(200, utils.Map{
			"guess":   guess,
			"player":  player,
			"new_elo": elo,
		})
	}
}
