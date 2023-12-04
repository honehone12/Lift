package handlers

import (
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
	Port uint16
}

func NextPort(c echo.Context) error {
	ctx, err := context.FromEchoContext(c)
	if err != nil {
		return errres.ServerError(err, c.Logger())
	}

	b := ctx.Components.Brain()
	p, err := b.Launch()
	if err != nil {
		return errres.ServerError(err, c.Logger())
	}

	return c.JSON(http.StatusOK, NextPortResponse{
		Port: p.Number(),
	})
}

type NextBackfillPortResponse struct {
}

func NextBackfillPort(c echo.Context) error {
	return errres.NotInService(c.Logger())
}
