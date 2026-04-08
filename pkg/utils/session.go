package utils

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

const (
	SessionCookieName = "rank_guessr_session"
)

func GetSessionCookie(ctx *echo.Context) (string, error) {
	cookie, err := ctx.Cookie(SessionCookieName)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

func SetSessionCookie(ctx *echo.Context, sessionId string) {
	cookie := &http.Cookie{
		Path:     "/",
		HttpOnly: true,
		Value:    sessionId,
		Name:     SessionCookieName,
		SameSite: http.SameSiteLaxMode,
	}

	ctx.SetCookie(cookie)
}

func UnsetSessionCookie(ctx *echo.Context) {
	cookie := &http.Cookie{
		Path:     "/",
		HttpOnly: true,
		Value:    "",
		MaxAge:   -1,
		Name:     SessionCookieName,
		SameSite: http.SameSiteLaxMode,
	}

	ctx.SetCookie(cookie)
}
