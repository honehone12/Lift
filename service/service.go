package service

import (
	"lift/gsmap/gsparams"
	"lift/server"
	"lift/server/context"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

const (
	ServiceName    = "LiftService"
	ServiceVersion = "0.0.1"

	ServerListenAt = "127.0.0.1:9990"

	MonitoringTimeout  = time.Second * 5
	InitialGSProcesses = 1
	GSProcessName      = "dummy"
	GSListenAddress    = "127.0.0.1"
)

func Run() {
	e := echo.New()
	c := context.NewComponents(
		context.NewMetadata(ServiceName, ServiceVersion),
	)
	p := server.NewServerParams(ServerListenAt, log.DEBUG)
	s := server.NewServer(e, c, p)

	errCh := s.Run()

	c.GSMap().Launch(gsparams.NewGSParams(
		GSProcessName,
		uuid.New(),
		GSListenAddress,
		"7777",
		MonitoringTimeout,
	), e.Logger)

	e.Logger.Fatal(<-errCh)
}
