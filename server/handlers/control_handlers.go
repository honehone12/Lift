package handlers

import (
	"lift/server/context"
	"lift/server/errres"
	"net/http"

	"github.com/labstack/echo/v4"
)

func ControlGSInfo(c echo.Context) error {
	ctx, err := context.FromEchoContext(c)
	if err != nil {
		return errres.ServerError(err, c.Logger())
	}

	gsMap := ctx.Components.GSMap()
	info, err := gsMap.UnsortedInfo()
	if err != nil {
		return errres.ServerError(err, c.Logger())
	}

	return c.JSON(http.StatusOK, info)
}

func ControlPortInfo(c echo.Context) error {
	ctx, err := context.FromEchoContext(c)
	if err != nil {
		return errres.ServerError(err, c.Logger())
	}

	info, err := ctx.Brain().PortMan().Info()
	if err != nil {
		return errres.ServerError(err, c.Logger())
	}

	return c.JSON(http.StatusOK, info)
}
