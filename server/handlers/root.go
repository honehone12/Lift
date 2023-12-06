package handlers

import (
	"lift/gsmap/gsinfo"
	"lift/server/context"
	"lift/server/errres"
	"net/http"

	"github.com/labstack/echo/v4"
)

type RootResponse struct {
	Name    string
	Version string
}

func Root(c echo.Context) error {
	ctx, err := context.FromEchoContext(c)
	if err != nil {
		return errres.ServerError(err, c.Logger())
	}

	m := ctx.Metadata()
	return c.JSON(http.StatusOK, &RootResponse{
		Name:    m.Name(),
		Version: m.Version(),
	})
}

type NextPortResponse struct {
	GsPort gsinfo.GSPort
}

func NextPort(c echo.Context) error {
	ctx, err := context.FromEchoContext(c)
	if err != nil {
		return errres.ServerError(err, c.Logger())
	}

	p, err := ctx.Components.Brain().Launch()
	if err != nil {
		return errres.ServerError(err, c.Logger())
	}

	return c.JSON(http.StatusOK, NextPortResponse{GsPort: *p})
}

type BackfillPortResponse struct {
	List []gsinfo.GSBackfillPort
}

func BackfillPort(c echo.Context) error {
	ctx, err := context.FromEchoContext(c)
	if err != nil {
		return errres.ServerError(err, c.Logger())
	}

	backfillList, err := ctx.Brain().BackfillList()
	if err != nil {
		return errres.ServerError(err, c.Logger())
	}

	return c.JSON(http.StatusOK, BackfillPortResponse{List: backfillList})
}
