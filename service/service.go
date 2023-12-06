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
	LogLevel = log.INFO

	ServiceName    = "LiftService"
	ServiceVersion = "0.0.1"
	ServerListenAt = "127.0.0.1:9990"

	InitialGSProcesses   = 0
	GSProcessName        = "dummy"
	GSListenAddress      = "127.0.0.1"
	GSMonitoringTimeout  = time.Second * 5
	GSConnectionCapacity = 2

	PortCapacity  = 100
	PortStartFrom = 7777

	BrainLoopInterval = time.Second * 10
	BrainMinimumWait  = time.Second * 10
)

func Run() {
	e := echo.New()
	gsm := gsmap.NewGSMap(e.Logger)
	b, err := brain.NewBrain(
		&brain.BrainParams{
			GSProcessName:        GSProcessName,
			GSListenAddress:      GSListenAddress,
			GSMonitoringTimeout:  GSMonitoringTimeout,
			GSConnectionCapacity: GSConnectionCapacity,
			PortParams: portman.PortManParams{
				InitialCapacity: PortCapacity,
				StartFrom:       PortStartFrom,
			},
			LoopInterval:        BrainLoopInterval,
			MinimumWaitForClose: BrainMinimumWait,
		},
		gsm,
		e.Logger,
	)
	if err != nil {
		e.Logger.Fatal(err)
	}

	s := server.NewServer(e,
		context.NewComponents(
			context.NewMetadata(ServiceName, ServiceVersion),
			gsm,
			b,
		),
		server.NewServerParams(ServerListenAt, LogLevel),
	)
	errCh := s.Run()

	for i := 0; i < InitialGSProcesses; i++ {
		if _, err := b.Launch(); err != nil {
			e.Logger.Fatal(err)
		}
	}

	e.Logger.Fatal(<-errCh)
}
