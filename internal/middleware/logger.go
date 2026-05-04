package middleware

import (
	"log/slog"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func RequestLogger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:    true,
		LogURI:       true,
		LogRequestID: true,
		LogLatency:   true,
		HandleError:  true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				slog.LogAttrs(c.Request().Context(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("reqid", v.RequestID),
					slog.String("latency", v.Latency.String()),
				)

				return nil
			}

			slog.LogAttrs(c.Request().Context(), slog.LevelError, "ERROR",
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.String("reqid", v.RequestID),
				slog.String("err", v.Error.Error()),
				slog.String("latency", v.Latency.String()),
			)

			return nil
		},
	})
}
