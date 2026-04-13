package middleware

import (
	"log/slog"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rankguessr/api/pkg/utils"
)

func RequestLogger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:    true,
		LogURI:       true,
		LogRequestID: true,
		LogLatency:   true,
		HandleError:  true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			session, err := utils.GetSession(c)

			if v.Error == nil {
				slog.LogAttrs(c.Request().Context(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("reqid", v.RequestID),
					slog.String("latency", v.Latency.String()),
				)
			} else if err != nil {
				slog.LogAttrs(c.Request().Context(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("reqid", v.RequestID),
					slog.String("err", v.Error.Error()),
					slog.String("latency", v.Latency.String()),
				)
			} else {
				slog.LogAttrs(c.Request().Context(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.Int("user_id", session.User.OsuID),
					slog.String("reqid", v.RequestID),
					slog.String("err", v.Error.Error()),
					slog.String("latency", v.Latency.String()),
				)
			}

			return nil
		},
	})
}
