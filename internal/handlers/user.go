package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/osuapi"
	"github.com/rankguessr/api/pkg/utils"
)

func UserGetRoomsData(rooms service.Rooms, client *osuapi.Client, guesses service.Guess) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return echo.ErrUnauthorized.Wrap(err)
		}

		var room *domain.Room
		found, err := rooms.FindByUserUnguessed(ctx, session.User.OsuID)
		if err == nil {
			room = &found
		}

		score, err := rooms.GetScore(ctx, session.AccessToken, room.ScoreID)
		if err != nil {
			err := rooms.DeleteById(ctx, room.ID)
			if err != nil {
				return err
			}

			return echo.ErrNotFound.Wrap(err)
		}

		if room.ClosesAt.Before(time.Now()) {
			player, err := client.GetUser(ctx, session.AccessToken, score.User.ID)
			if err != nil {
				return echo.ErrInternalServerError.Wrap(err)
			}

			_, _, err = guesses.Create(
				ctx, session.User.OsuID, domain.GuessCreate{
					PlayerID:     player.ID,
					Guess:        0,
					ScoreID:      room.ScoreID,
					BeatmapID:    score.BeatmapID,
					BeatmapSetID: score.Beatmap.BeatmapSetId,
					ActualRank:   player.Statistics.GlobalRank,
				},
			)
			if err != nil {
				return echo.ErrInternalServerError.Wrap(err)
			}

			err = rooms.DeleteById(ctx, room.ID)
			if err != nil {
				return echo.ErrInternalServerError.Wrap(err)
			}

			room = nil
		}

		latest, err := guesses.FindByUser(ctx, session.User.OsuID, 6)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		if room == nil {
			return c.JSON(http.StatusOK, utils.Map{
				"latest": latest,
			})
		}

		// TODO: add a mapper from score to anonymized
		return c.JSON(http.StatusOK, utils.Map{
			"room": utils.Map{
				"id":        room.ID,
				"closes_at": room.ClosesAt,
				"score": utils.Map{
					"pp":         score.PP,
					"mods":       score.Mods,
					"accuracy":   score.Accuracy,
					"beatmapset": score.BeatmapSet,
					"beatmap":    score.Beatmap,
					"statistics": score.Statistics,
				},
			},
			"latest": latest,
		})
	}
}
