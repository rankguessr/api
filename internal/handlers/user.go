package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/pkg/utils"
)

func UserGetCurrentRoom(rooms service.Rooms) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return echo.ErrUnauthorized.Wrap(err)
		}

		room, err := rooms.FindByUser(ctx, session.User.OsuID, session.AccessToken)
		if err != nil {
			return c.JSON(http.StatusOK, utils.Map{
				"room": nil,
			})
		}

		return c.JSON(http.StatusOK, utils.Map{
			"room": utils.Map{
				"id":        room.ID,
				"kind":      room.Kind,
				"closes_at": room.ClosesAt,
				"score": utils.Map{
					"pp":         room.Score.PP,
					"accuracy":   room.Score.Accuracy,
					"beatmapset": room.Score.BeatmapSet,
					"beatmap":    room.Score.Beatmap,
					"statistics": room.Score.Statistics,
					"mods":       room.Score.ModsAcronyms(),
				},
			},
		})
	}
}

func UserGetGuesses(guesses service.Guess) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return echo.ErrUnauthorized.Wrap(err)
		}

		limit, err := strconv.Atoi(c.QueryParam("limit"))
		if err != nil || limit <= 0 {
			limit = 10
		}

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		if limit > 50 {
			return echo.ErrBadRequest.Wrap(utils.ErrLimitExceeded)
		}

		result, err := guesses.FindByUser(ctx, session.User.OsuID, limit, page)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		return c.JSON(http.StatusOK, result)
	}
}
