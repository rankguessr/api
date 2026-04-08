package utils

import (
	"github.com/labstack/echo/v5"
	"github.com/rankguessr/api/pkg/domain"
)

const sessionCtxKey = "session"

func SetSession(ctx *echo.Context, session domain.SessionExtended) {
	ctx.Set(sessionCtxKey, session)
}

func GetSession(ctx *echo.Context) (domain.SessionExtended, error) {
	session, err := echo.ContextGet[domain.SessionExtended](ctx, sessionCtxKey)
	if err != nil {
		return domain.SessionExtended{}, err
	}

	return session, nil
}
