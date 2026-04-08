package utils

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

const (
	AuthStateCookieName = "auth_state"
)

func SetAuthStateCookie(ctx *echo.Context, domain, value string) {
	cookie := &http.Cookie{
		Path:     "/",
		HttpOnly: true,
		Value:    value,
		Domain:   domain,
		Name:     AuthStateCookieName,
		SameSite: http.SameSiteLaxMode,
	}

	ctx.SetCookie(cookie)
}

func UnsetAuthStateCookie(ctx *echo.Context, domain string) {
	cookie := &http.Cookie{
		Path:     "/",
		HttpOnly: true,
		Value:    "",
		MaxAge:   -1,
		Domain:   domain,
		Name:     AuthStateCookieName,
		SameSite: http.SameSiteLaxMode,
	}

	ctx.SetCookie(cookie)
}

func ReadAuthStateCookie(ctx *echo.Context) (string, error) {
	cookie, err := ctx.Cookie(AuthStateCookieName)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}
