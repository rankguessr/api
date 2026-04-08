package middleware

import (
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/pkg/osuapi"
	"github.com/rankguessr/api/pkg/utils"
)

func Session(api *osuapi.Client, sessions service.Sessions) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx := c.Request().Context()
			sessionId, err := utils.GetSessionCookie(c)
			if err != nil {
				return echo.ErrUnauthorized.Wrap(err)
			}

			session, err := sessions.FindWithUser(ctx, sessionId)
			if err != nil {
				return echo.ErrUnauthorized.Wrap(err)
			}

			if time.Until(session.ExpiresAt) <= time.Second*30 {
				token, err := api.TokenRefresh(ctx, session.RefreshToken)
				if err != nil {
					return echo.ErrUnauthorized.Wrap(err)
				}

				expiresAt := token.ExpiresAt()
				err = sessions.UpdateTokens(ctx, sessionId, token.AccessToken, token.RefreshToken, expiresAt)
				if err != nil {
					return echo.ErrInternalServerError.Wrap(err)
				}

				session.AccessToken = token.AccessToken
				session.RefreshToken = token.RefreshToken
				session.ExpiresAt = expiresAt
			}

			utils.SetSession(c, session)

			return next(c)
		}
	}
}
