package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/rankguessr/api/internal/service"
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

		room, roomErr := rooms.FindByUser(ctx, session.User.OsuID, session.AccessToken)

		latest, err := guesses.FindByUser(ctx, session.User.OsuID, 6)
		if err != nil {
			return err
		}

		if roomErr != nil {
			return c.JSON(http.StatusOK, utils.Map{
				"room":   nil,
				"latest": latest,
			})
		}

		// TODO: add a mapper from score to anonymized
		return c.JSON(http.StatusOK, utils.Map{
			"room": utils.Map{
				"id":        room.ID,
				"closes_at": room.ClosesAt,
				"score": utils.Map{
					"pp":         room.Score.PP,
					"mods":       room.Score.Mods,
					"accuracy":   room.Score.Accuracy,
					"beatmapset": room.Score.BeatmapSet,
					"beatmap":    room.Score.Beatmap,
					"statistics": room.Score.Statistics,
				},
			},
			"latest": latest,
		})
	}
}
