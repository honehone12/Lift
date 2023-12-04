package gs

import (
	"bytes"
	"lift/gsmap/gsinfo"
	"lift/gsmap/gsparams"
	"lift/gsmap/gsprocess"
	"lift/gsmap/monitor"
	"lift/logger"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type GS struct {
	params  *gsparams.GSParams
	process *gsprocess.GSProcess
	conn    *websocket.Conn
	logger  logger.Logger

	summary gsinfo.MonitoringSummary

	onGSClosed  func() error
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
		summary:     gsinfo.MonitoringSummary{},
		closingWait: sync.WaitGroup{},
		closeCh:     make(chan bool),
	}, nil
}

func (gs *GS) StartProcess(onGSClosed func() error) error {
	err := gs.process.Start(func() {
		gs.closingWait.Done()
		gs.closeCh <- true
	})
	if err != nil {
		return err
	}

	gs.summary.TimeStarted = time.Now()
	gs.summary.MonitoringStatus = gsinfo.MonitoringStatusNotConnected
	gs.onGSClosed = onGSClosed
	gs.closingWait.Add(1)
	gs.logger.Info(gs.params.LogWithId("gs successfully started"))
	return nil
}

func (gs *GS) EndProcess() {
	gs.process.Close()
}

func (gs *GS) Info() gsinfo.GSInfo {
	return gsinfo.GSInfo{
		ID:            gs.params.UuidString(),
		Port:          gs.params.Port().Number(),
		ProcessStatus: gs.process.Status(),
		Summary:       gs.summary,
	}
}

func (gs *GS) Established() bool {
	return gs.conn != nil
}

func (gs *GS) StartListen(conn *websocket.Conn) {
	if conn == nil || gs.conn != nil {
		return
	}

	gs.conn = conn
	gs.summary.TimeEstablished = time.Now()
	gs.summary.MonitoringStatus = gsinfo.MonitoringStatusOK
	gs.closingWait.Add(1)
	go gs.listen()
	go gs.wait()
}

func (gs *GS) wait() {
	gs.closingWait.Wait()
	gs.summary.MonitoringStatus = gsinfo.MonitoringStatusClosed
	gs.summary.TimeClosed = time.Now()
	gs.summary.ConnectionCount = -1
	gs.summary.SessionCount = -1
	gs.summary.ActiveSessionCount = -1
	gs.logger.Info(gs.params.LogWithId("gs successfully closed"))
}

func (gs *GS) recoverListen() {
	if r := recover(); r != nil {
		gs.logger.Warn(gs.params.LogWithId("recovering listening goroutine"))
		go gs.listen()
	}
}

func (gs *GS) listen() {
	defer gs.recoverListen()

LOOP:
	for {
		select {
		case <-gs.closeCh:
			if gs.conn != nil {
				defer gs.conn.Close()
			}
			err := gs.onGSClosed()
			if err != nil {
				gs.logger.Warnf(gs.params.LogWithId("error on close"), err)
			}
			break LOOP
		default:
			if gs.summary.MonitoringStatus != gsinfo.MonitoringStatusOK {
				continue
			}

			if err := gs.conn.SetReadDeadline(gs.params.NextMonitoringTimeout()); err != nil {
				gs.logger.Panic(gs.params.LogWithId(err.Error()))
			}

			m := monitor.MonitoringMessage{}
			if err := gs.conn.ReadJSON(&m); err != nil {
				gs.logger.Errorf(gs.params.LogWithId(
					"errror: %s, waiting for closing listening goroutine"),
					err.Error(),
				)

				gs.summary.ConnectionCount = -1
				gs.summary.SessionCount = -1
				gs.summary.ActiveSessionCount = -1
				gs.summary.MonitoringStatus = gsinfo.MonitoringStatusConnectionError
				continue
			}

			if !bytes.Equal(m.GuidRaw, gs.params.UuidRaw()) {
				gs.logger.Warn(gs.params.LogWithId("received broken uuid"))
				gs.summary.MonitoringStatus = gsinfo.MonitoringStatusConnectionError
				continue
			}

			if m.ErrorCode == monitor.ErrorFatal {
				gs.logger.Error(gs.params.LogWithId(string(m.ErrorUtf8)))

				gs.summary.ConnectionCount = -1
				gs.summary.SessionCount = -1
				gs.summary.ActiveSessionCount = -1
				gs.summary.MonitoringStatus = gsinfo.MonitoringStatusError
				continue
			} else if m.ErrorCode == monitor.ErrorWarn {
				gs.logger.Warn(gs.params.LogWithId(string(m.ErrorUtf8)))
			}

			gs.logger.Debugf(gs.params.LogWithId("%#v"), m)
			gs.summary.ConnectionCount = m.ConnectionCount
			gs.summary.SessionCount = m.SessionCount
			gs.summary.ActiveSessionCount = m.ActiveSessionCount
		}
	}

	gs.logger.Debug(gs.params.LogWithId("listening goroutine successfully closed"))
	gs.closingWait.Done()
}
