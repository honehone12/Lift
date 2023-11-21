package service

import (
	"lift/server"
	"lift/server/context"

	"github.com/labstack/echo/v4"
)

func Run() {
	e := echo.New()
	server.NewServer(e,
		context.NewComponents(
			context.NewMetadata("LiftService", "0.0.1"),
		),
		"127.0.0.1:9990",
	).Run()
}
