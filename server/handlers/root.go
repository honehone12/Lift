package handlers

import (
	"lift/brain"
	"lift/gsmap/gsinfo"
	"lift/server/context"
	"lift/server/errres"
	"net/http"
	"strconv"

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
	idxStr := c.Param("index")
	idx, err := strconv.ParseInt(idxStr, 10, 64)
	if err != nil {
		return errres.BadRequest(err, c.Logger())
	}

	ctx, err := context.FromEchoContext(c)
	if err != nil {
		return errres.ServerError(err, c.Logger())
	}
	b := ctx.Brain()

	p, err := b.Launch(int(idx))
	if err == brain.ErrorIndexOutOfRange {
		return errres.BadRequest(err, c.Logger())
	} else if err != nil {
		return errres.ServerError(err, c.Logger())
	}

	return c.JSON(http.StatusOK, NextPortResponse{GsPort: *p})
}

type BackfillPortResponse struct {
	List []gsinfo.GSBackfillPort
}

func BackfillPort(c echo.Context) error {
	idxStr := c.Param("index")
	idx, err := strconv.ParseInt(idxStr, 10, 64)
	if err != nil {
		return errres.BadRequest(err, c.Logger())
	}

	ctx, err := context.FromEchoContext(c)
	if err != nil {
		return errres.ServerError(err, c.Logger())
	}
	b := ctx.Brain()

	backfillList, err := b.BackfillList(int(idx))
	if err == brain.ErrorIndexOutOfRange {
		return errres.BadRequest(err, c.Logger())
	} else if err != nil {
		return errres.ServerError(err, c.Logger())
	}

	return c.JSON(http.StatusOK, BackfillPortResponse{List: backfillList})
}
