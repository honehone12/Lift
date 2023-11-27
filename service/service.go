package service

import (
	"lift/brain"
	"lift/brain/portman"
	"lift/gsmap"
	"lift/server"
	"lift/server/context"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

const (
	ServiceName    = "LiftService"
	ServiceVersion = "0.0.1"
	ServerListenAt = "127.0.0.1:9990"

	GSMonitoringTimeout = time.Second * 5
	InitialGSProcesses  = 1
	GSProcessName       = "dummy"
	GSListenAddress     = "127.0.0.1"

	PortCapacity  = 100
	PortStartFrom = 7777
)

func Run() {
	e := echo.New()
	gsm := gsmap.NewGSMap(e.Logger)
	b, err := brain.NewBrain(&brain.BrainParams{
		GSProcessName:       GSProcessName,
		GSListenAddress:     GSListenAddress,
		GSMonitoringTimeout: GSMonitoringTimeout,
		PortParams: portman.PortManParams{
			InitialCapacity: PortCapacity,
			StartFrom:       PortStartFrom,
		},
	}, gsm)
	if err != nil {
		e.Logger.Fatal(err)
	}

	s := server.NewServer(e,
		context.NewComponents(
			context.NewMetadata(ServiceName, ServiceVersion),
			gsm,
			b,
		),
		server.NewServerParams(ServerListenAt, log.DEBUG),
	)
	errCh := s.Run()

	if err = b.Launch(InitialGSProcesses); err != nil {
		e.Logger.Fatal(err)
	}

	e.Logger.Fatal(<-errCh)
}
