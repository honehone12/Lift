package service

import (
	"encoding/json"
	"flag"
	"io"
	"lift/brain"
	"lift/brain/portman"
	"lift/gsmap"
	"lift/server"
	"lift/server/context"
	"lift/setting"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func parseFlags() string {
	fileName := flag.String("s", "setting.json", "setting file name")
	flag.Parse()
	return *fileName
}

func loadSetting(file string) (*setting.Setting, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	s := &setting.Setting{}
	err = json.Unmarshal(b, s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func Run() {
	e := echo.New()
	fileName := parseFlags()
	setting, err := loadSetting(fileName)
	if err != nil {
		e.Logger.Fatal(err)
	}

	gsm := gsmap.NewGSMap(e.Logger)
	b, err := brain.NewBrain(
		&brain.BrainParams{
			GSExecutables:    setting.GSExecutables,
			GSListenAddress:  setting.GSListenAddress,
			GSMessageTimeout: time.Second * time.Duration(setting.GSMessageTimeoutSec),
			PortParams: portman.PortManParams{
				InitialCapacity: setting.PortCapacity,
				StartFrom:       setting.PortStartFrom,
			},
			LoopInterval:        time.Second * time.Duration(setting.BrainIntervalSec),
			MinimumWaitForClose: time.Second * time.Duration(setting.BrainMinimumWaitSec),
		},
		gsm,
		e.Logger,
	)
	if err != nil {
		e.Logger.Fatal(err)
	}

	s := server.NewServer(e,
		context.NewComponents(
			context.NewMetadata(setting.ServiceName, setting.ServiceVersion),
			gsm,
			b,
		),
		server.NewServerParams(setting.ServiceListenAt, log.Lvl(setting.LogLevel)),
	)
	errCh := s.Run()

	e.Logger.Fatal(<-errCh)
}
