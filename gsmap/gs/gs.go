package gs

import (
	"bytes"
	"lift/gsmap/gsinfo"
	"lift/gsmap/gsparams"
	"lift/gsmap/gsprocess"
	"lift/gsmap/monitor"
	"lift/logger"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type GS struct {
	params  *gsparams.GSParams
	process *gsprocess.GSProcess
	conn    *websocket.Conn
	logger  logger.Logger

	timeStarted         *time.Time
	timeEstablished     *atomic.Pointer[time.Time]
	timeLastCommunicate *atomic.Pointer[time.Time]

	lastConnectionCount    *atomic.Int64
	lastSessionCount       *atomic.Int64
	lastActiveSessionCount *atomic.Int64

	fatal *atomic.Bool

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
		params:                 params,
		process:                process,
		logger:                 logger,
		timeStarted:            nil,
		timeEstablished:        &atomic.Pointer[time.Time]{},
		timeLastCommunicate:    &atomic.Pointer[time.Time]{},
		lastConnectionCount:    &atomic.Int64{},
		lastSessionCount:       &atomic.Int64{},
		lastActiveSessionCount: &atomic.Int64{},
		fatal:                  &atomic.Bool{},
		closingWait:            sync.WaitGroup{},
		closeCh:                make(chan bool),
	}, nil
}

func (gs *GS) StartProcess(onGSClosed func() error) error {
	err := gs.process.Start(func() {
		gs.closeCh <- true
		gs.closingWait.Done()
	})
	if err != nil {
		return err
	}

	now := time.Now()
	gs.timeStarted = &now
	gs.onGSClosed = onGSClosed
	gs.closingWait.Add(2)
	go gs.listen()
	go gs.wait()
	gs.logger.Info(gs.params.LogWithId("gs successfully started"))
	return nil
}

func (gs *GS) EndProcess() {
	gs.process.Close()
}

func (gs *GS) Info() gsinfo.GSInfo {
	i := gsinfo.GSInfo{
		ID:   gs.params.UuidString(),
		Port: gs.params.Port().Number(),
		Summary: gsinfo.MonitoringSummary{
			ConnectionCount:    gs.lastConnectionCount.Load(),
			SessionCount:       gs.lastSessionCount.Load(),
			ActiveSessionCount: gs.lastActiveSessionCount.Load(),
		},
		Fatal: gs.fatal.Load(),
	}
	var ptr *time.Time
	if ptr = gs.timeStarted; ptr != nil {
		i.Summary.TimeStarted = *ptr
	}
	if ptr = gs.timeEstablished.Load(); ptr != nil {
		i.Summary.TimeEstablished = *ptr
	}
	if ptr = gs.timeLastCommunicate.Load(); ptr != nil {
		i.Summary.TimeLastCommunicate = *ptr
	}

	return i
}

func (gs *GS) Established() bool {
	return gs.conn != nil
}

func (gs *GS) StartListen(conn *websocket.Conn) {
	if conn == nil || gs.conn != nil {
		return
	}

	now := time.Now()
	gs.timeEstablished.Store(&now)
	gs.conn = conn
}

func (gs *GS) wait() {
	gs.closingWait.Wait()
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

	connectionBroken := false

LOOP:
	for {
		select {
		case <-gs.closeCh:
			if gs.conn != nil {
				defer gs.conn.Close()
			}
			defer gs.onGSClosed()
			break LOOP
		default:
			if gs.conn == nil {
				continue
			}

			if connectionBroken {
				continue
			}

			if err := gs.conn.SetReadDeadline(gs.params.NextMonitoringTimeout()); err != nil {
				gs.logger.Panicf(gs.params.LogWithId(
					"%s: this means time settting is broken"),
					err.Error(),
				)
			}

			m := monitor.MonitoringMessage{}
			if err := gs.conn.ReadJSON(&m); err != nil {
				gs.logger.Errorf(gs.params.LogWithId(
					"errror: %s, waiting for closing listening goroutine"),
					err.Error(),
				)
				connectionBroken = true
				gs.fatal.Store(true)
				continue
			}

			if !bytes.Equal(m.GuidRaw, gs.params.UuidRaw()) {
				gs.logger.Warn(gs.params.LogWithId("received broken uuid"))
				connectionBroken = true
				gs.fatal.Store(true)
				continue
			}

			now := time.Now()
			gs.timeLastCommunicate.Store(&now)

			if m.ErrorCode == monitor.ErrorFatal {
				gs.logger.Error(gs.params.LogWithId(string(m.ErrorUtf8)))
				gs.fatal.Store(true)
				continue
			} else if m.ErrorCode == monitor.ErrorWarn {
				gs.logger.Warn(gs.params.LogWithId(string(m.ErrorUtf8)))
			}

			gs.logger.Debugf(gs.params.LogWithId("%#v"), m)
			gs.lastConnectionCount.Store(m.ConnectionCount)
			gs.lastSessionCount.Store(m.SessionCount)
			gs.lastActiveSessionCount.Store(m.ActiveSessionCount)
		}
	}

	gs.logger.Debug(gs.params.LogWithId("listening goroutine successfully closed"))
	gs.closingWait.Done()
}
