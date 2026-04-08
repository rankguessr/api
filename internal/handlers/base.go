package handlers

import (
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/pkg/utils"
)

func HealthCheck(ctx *echo.Context) error {
	return ctx.JSON(200, utils.Map{
		"status": "ok",
	})
}

func PublicStatsGet(guess service.Guess) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		count24h, err := guess.CountFromDate(ctx, time.Now().Add(-24*time.Hour))
		if err != nil {
			return c.JSON(500, utils.Map{
				"error": "failed to get stats",
			})
		}

		countGlobal, err := guess.CountFromDate(ctx, time.Unix(0, 0))
		if err != nil {
			return c.JSON(500, utils.Map{
				"error": "failed to get stats",
			})
		}

		best, err := guess.FindTopFromDate(ctx, time.Now().Add(-24*time.Hour), 10)
		if err != nil {
			return c.JSON(500, utils.Map{
				"error": "failed to get stats",
			})
		}

		latest, err := guess.FindLatest(ctx)
		if err != nil {
			return c.JSON(500, utils.Map{
				"error": "failed to get stats",
			})
		}

		return c.JSON(200, utils.Map{
			"best":         best,
			"latest":       latest,
			"count_24h":    count24h,
			"count_global": countGlobal,
		})
	}
}
