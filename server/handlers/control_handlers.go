package handlers

import (
	"lift/gsmap/gsinfo"
	"lift/server/context"
	"lift/server/errres"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ControlIndexResponse struct {
	List []gsinfo.GSClass
}

func ControlIndex(c echo.Context) error {
	ctx, err := context.FromEchoContext(c)
	if err != nil {
		return errres.ServerError(err, c.Logger())
	}

	return c.JSON(http.StatusOK, ControlIndexResponse{
		List: ctx.Brain().ExecutableList(),
	})
}

func ControlGSInfo(c echo.Context) error {
	ctx, err := context.FromEchoContext(c)
	if err != nil {
		return errres.ServerError(err, c.Logger())
	}

	info, err := ctx.GSMap().UnsortedInfo()
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
