package handlers

import (
	"log"
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

		latest, err := guesses.FindByUser(ctx, session.User.OsuID, 10)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		room, err := rooms.FindByUserUnguessed(ctx, session.User.OsuID)
		if err != nil {
			log.Println(room)
			return c.JSON(http.StatusOK, utils.Map{
				"room": nil,
			})
		}

		score, err := client.GetScore(ctx, session.AccessToken, room.ScoreID)
		if err != nil {
			err := rooms.DeleteById(ctx, room.ID)
			if err != nil {
				return echo.ErrInternalServerError.Wrap(err)
			}

			return c.JSON(http.StatusOK, utils.Map{
				"room":   nil,
				"latest": latest,
			})
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
					"beatmap":    score.Beatmap,
					"statistics": score.Statistics,
				},
			},
			"latest": latest,
		})
	}
}
