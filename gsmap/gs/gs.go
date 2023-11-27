package gs

import (
	"fmt"
	"lift/gsmap/gsparams"
	"lift/gsmap/gsprocess"
	"lift/gsmap/monitor"
	"lift/logger"
	"sync"

	"github.com/gorilla/websocket"
)

type GS struct {
	params  *gsparams.GSParams
	process *gsprocess.GSProcess
	conn    *websocket.Conn
	logger  logger.Logger

	closingWait sync.WaitGroup
	closeCh     chan bool
}

func NewGS(params *gsparams.GSParams, logger logger.Logger) (*GS, error) {
	process, err := gsprocess.NewGSProcess(params, logger)
	if err != nil {
		return nil, err
	}

	return &GS{
		params:      params,
		process:     process,
		logger:      logger,
		closingWait: sync.WaitGroup{},
		closeCh:     make(chan bool),
	}, nil
}

func (gs *GS) processDone() {
	gs.closingWait.Done()
	gs.closeCh <- true
}

func (gs *GS) StartProcess() error {
	err := gs.process.Start(gs.processDone)
	if err != nil {
		return err
	}

	gs.closingWait.Add(1)
	gs.logger.Info(gs.params.LogFormat("gs successfully started"))
	return nil
}

func (gs *GS) Established() bool {
	return gs.conn != nil
}

func (gs *GS) StartListen(conn *websocket.Conn) {
	if conn == nil || gs.conn != nil {
		return
	}

	gs.conn = conn
	gs.closingWait.Add(1)
	go gs.listen()
	go gs.wait()
}

func (gs *GS) wait() {
	gs.closingWait.Wait()
	gs.logger.Info(gs.params.LogFormat("gs successfully closed"))
}

func (gs *GS) recoverListen() {
	if r := recover(); r != nil {
		gs.logger.Warn(gs.params.LogFormat("recovering listening goroutine"))
		go gs.listen()
	}
}

func (gs *GS) listen() {
	defer gs.recoverListen()

LOOP:
	for {
		select {
		case <-gs.closeCh:
			break LOOP
		default:
			if gs.conn == nil {
				continue
			}

			if err := gs.conn.SetReadDeadline(gs.params.NextMonitoringTimeout()); err != nil {
				gs.logger.Panic(gs.params.LogFormat(err.Error()))
			}

			m := monitor.MonitoringMessage{}
			if err := gs.conn.ReadJSON(&m); err != nil {
				gs.logger.Error(gs.params.LogFormat(fmt.Sprintf(
					"errror: %s, waiting for closing listening goroutine", err.Error(),
				)))
				defer gs.conn.Close()
				gs.conn = nil
				continue
			}

			gs.logger.Infof(gs.params.LogFormat("%#v"), m)
		}
	}

	gs.logger.Debug(gs.params.LogFormat("listening goroutine successfully closed"))
	gs.closingWait.Done()
}
