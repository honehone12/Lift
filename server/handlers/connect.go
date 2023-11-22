package handlers

import (
	"errors"
	"lift/server/context"
	"lift/server/errres"

	"github.com/labstack/echo/v4"
)

type ConnectParam struct {
	ProcessId string `validate:"required,uuid4,min=36,max=36"`
}

var (
	ErrorDuplicatedConnection = errors.New("attempt to duplicate connection")
)

func Connect(c echo.Context) error {
	param := ConnectParam{
		ProcessId: c.Param("id"),
	}
	if err := c.Validate(&param); err != nil {
		return errres.BadRequest(err, c.Logger())
	}

	ctx, err := context.FromEchoContext(c)
	if err != nil {
		return errres.ServerError(err, c.Logger())
	}

	gs, err := ctx.GSMap().Item(param.ProcessId)
	if err != nil {
		return errres.BadRequest(err, c.Logger())
	}

	if gs.Established() {
		return errres.BadRequest(ErrorDuplicatedConnection, c.Logger())
	}

	conn, err := ctx.WebSocketUpgrader().Upgrade(
		c.Response(),
		c.Request(),
		nil,
	)
	if err != nil {
		return errres.ServerError(err, c.Logger())
	}

	gs.StartListen(conn)
	return nil
}
