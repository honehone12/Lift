package main

import (
	"lift/gs/gsparams"
	"lift/gs/gsprocess"
	"lift/server"
	"lift/server/context"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func RunServer(e *echo.Echo) {
	server.NewServer(e,
		context.NewComponents(
			context.NewMetadata("ThisIsTheTest", "99.99.99"),
		),
		"127.0.0.1:9990",
	).Run()
}

func main() {
	e := echo.New()

	gsParams := gsparams.NewGSParams(
		"dummy",
		uuid.New(),
		"127.0.0.1",
		"7777",
	)
	gs, err := gsprocess.NewGSProcess(gsParams, e.Logger)
	if err != nil {
		panic(err)
	}

	err = gs.Start()
	if err != nil {
		panic(err)
	}

	RunServer(e)
}
