package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/pkg/cache"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/utils"
	"github.com/redis/go-redis/v9"
)

func HealthCheck(ctx *echo.Context) error {
	return ctx.JSON(http.StatusOK, utils.Map{
		"status": "ok",
	})
}

func PublicStatsGet(guess service.Guess, users service.User, rdb *redis.Client) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		stats, err := cache.GetStats(rdb, ctx)
		if err == nil {
			return c.JSON(200, stats)
		}

		count24h, err := guess.CountFromDate(ctx, time.Now().Add(-24*time.Hour))
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		countGlobal, err := guess.CountFromDate(ctx, time.Unix(0, 0))
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		bestGuesses, err := guess.FindTopFromDate(ctx, time.Now().Add(-24*time.Hour), 10)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		topUsers, err := users.FindTop(ctx, 15)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		stats = domain.Stats{
			TopUsers:    topUsers,
			Count24h:    count24h,
			CountGlobal: countGlobal,
			Best:        bestGuesses,
		}

		err = cache.SetStats(rdb, ctx, stats)
		if err != nil {
			slog.ErrorContext(ctx, "failed to set stats cache")
		}

		return c.JSON(200, stats)
	}
}
