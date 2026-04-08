package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/rankguessr/api/internal/service"
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

func UserGetLatest(user service.User, guesses service.Guess) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return echo.ErrUnauthorized.Wrap(err)
		}

		latest, err := guesses.FindByUser(ctx, session.User.OsuID, 15)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		return c.JSON(http.StatusOK, latest)
	}
}
