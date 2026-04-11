package handlers

import (
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/rankguessr/api/internal/config"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/pkg/osuapi"
	"github.com/rankguessr/api/pkg/utils"
)

func AuthMe(user service.User) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
		session, err := utils.GetSession(ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, utils.Map{
				"error": "invalid session token",
			})
		}

		u, err := user.FindByOsuID(ctx.Request().Context(), session.User.OsuID)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		return ctx.JSON(200, u)
	}
}

func AuthLogin(cfg *config.Config) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
		state := uuid.New().String()

		params := url.Values{}
		params.Add("state", state)
		params.Add("response_type", "code")
		params.Add("scope", "public identify")
		params.Add("client_id", cfg.OsuClientID)
		params.Add("redirect_uri", cfg.AppURL+"/auth/callback")

		utils.SetAuthStateCookie(ctx, cfg.WebDomain(), state)

		return ctx.Redirect(302, "https://osu.ppy.sh/oauth/authorize?"+params.Encode())
	}
}

func AuthCallback(cfg *config.Config, client *osuapi.Client, users service.User, sessions service.Sessions) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		code := c.QueryParam("code")
		state := c.QueryParam("state")

		if code == "" || state == "" {
			return c.JSON(http.StatusBadRequest, utils.Map{
				"error": "missing code or state",
			})
		}

		authState, err := utils.ReadAuthStateCookie(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, utils.Map{
				"error": "invalid auth state",
			})
		}

		if authState != state {
			return c.JSON(http.StatusUnauthorized, utils.Map{
				"error": "auth state mismatch",
			})
		}

		token, err := client.ExchangeToken(ctx, code)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		me, err := client.GetMe(ctx, token.AccessToken)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		err = users.Upsert(ctx, me.ID, me.Username, me.AvatarURL, me.CountryCode)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		session, err := sessions.Create(ctx, me.ID, token.AccessToken, token.RefreshToken, token.ExpiresAt())
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		utils.UnsetAuthStateCookie(c, cfg.WebDomain())
		utils.SetSessionCookie(c, cfg.WebDomain(), session.ID)

		return c.Redirect(302, cfg.WebURL)
	}
}

func AuthLogout(cfg *config.Config) echo.HandlerFunc {
	return func(c *echo.Context) error {
		utils.UnsetSessionCookie(c, cfg.WebDomain())

		return c.Redirect(http.StatusFound, cfg.WebURL)
	}
}
