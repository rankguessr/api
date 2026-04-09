package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/pkg/osuapi"
	"github.com/rankguessr/api/pkg/utils"
)

func UserGetMe(user service.User) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return echo.ErrUnauthorized.Wrap(err)
		}

		u, err := user.FindByOsuID(ctx, session.User.OsuID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, utils.Map{
				"error": "failed to get user",
			})
		}

		return c.JSON(http.StatusOK, utils.Map{
			"osu_id":       u.OsuID,
			"username":     u.Username,
			"avatar_url":   u.AvatarURL,
			"country_code": u.CountryCode,
		})
	}
}

func UserGetCurrentRoom(rooms service.Rooms, client *osuapi.Client) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return echo.ErrUnauthorized.Wrap(err)
		}

		room, err := rooms.FindByUser(ctx, session.User.OsuID)
		if err != nil {
			return c.JSON(http.StatusOK, utils.Map{
				"room": nil,
			})
		}

		score, err := client.GetScore(ctx, session.AccessToken, room.ScoreID)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		// TODO: add a mapper from score to anonymized
		return c.JSON(http.StatusOK, utils.Map{
			"room": utils.Map{
				"id": room.ID,
				"score": utils.Map{
					"pp":         score.PP,
					"mods":       score.Mods,
					"accuracy":   score.Accuracy,
					"beatmapset": score.BeatmapSet,
					"beatmap":    score.BeatmapExtended,
					"statistics": score.Statistics,
				},
			},
		})
	}
}

func UserGetLatest(user service.User, guesses service.Guess) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return echo.ErrUnauthorized.Wrap(err)
		}

		latest, err := guesses.FindByUser(ctx, session.User.OsuID, 10)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		return c.JSON(http.StatusOK, latest)
	}
}
