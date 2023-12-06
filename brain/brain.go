package brain

import (
	"lift/brain/portman"
	"lift/brain/portman/port"
	"lift/gsmap"
	"lift/gsmap/gs"
	"lift/gsmap/gsinfo"
	"lift/gsmap/gsparams"
	"lift/logger"
	"time"

	libuuid "github.com/google/uuid"
)

type BrainParams struct {
	GSProcessName        string
	GSListenAddress      string
	GSMonitoringTimeout  time.Duration
	GSConnectionCapacity int

	PortParams          portman.PortManParams
	LoopInterval        time.Duration
	MinimumWaitForClose time.Duration
}

type Brain struct {
	params  *BrainParams
	portMan *portman.PortMan

	lessFunc func(r, l *gsinfo.MonitoringSummary) bool

	gsMap   *gsmap.GSMap
	logger  logger.Logger
	ticker  *time.Ticker
	closeCh chan bool
}

func GenerateId() [16]byte {
	return libuuid.New()
}

func NewBrain(
	params *BrainParams,
	lessFunc func(r, l *gsinfo.MonitoringSummary) bool,
	gsMap *gsmap.GSMap,
	logger logger.Logger,
) (*Brain, error) {
	pm, err := portman.NewPortMan(params.PortParams)
	if err != nil {
		return nil, err
	}

	b := &Brain{
		params:   params,
		portMan:  pm,
		lessFunc: lessFunc,
		gsMap:    gsMap,
		logger:   logger,
		ticker:   time.NewTicker(params.LoopInterval),
		closeCh:  make(chan bool),
	}

	go b.brainMain()

	return b, nil
}

func (b *Brain) PortMan() *portman.PortMan {
	return b.portMan
}

func (b *Brain) Launch() (port.Port, error) {
	p, err := b.portMan.Next()
	if err != nil {
		return port.Port{}, err
	}

	param := gsparams.NewGSParams(
		b.params.GSProcessName,
		GenerateId(),
		b.params.GSListenAddress,
		p,
		b.params.GSMonitoringTimeout,
	)
	gs, err := gs.NewGS(param, b.logger)
	if err != nil {
		return port.Port{}, err
	}

	id := param.UuidString()
	if gs.StartProcess(func() error {
		if err := b.portMan.Return(p); err != nil {
			return err
		}

		b.gsMap.Remove(id)
		b.logger.Debugf(
			"process id: %s removed from gsmap, returned port: %d",
			id, p.Number(),
		)
		return nil
	}); err != nil {
		return port.Port{}, err
	}

	b.gsMap.Add(id, gs)
	return p, nil
}

func (b *Brain) Shutdown(id string) error {
	gs, err := b.gsMap.Item(id)
	if err != nil {
		return err
	}

	b.logger.Debugf("brain start closing process id: %s", id)
	gs.EndProcess()
	return nil
}

func (b *Brain) recoverBrainMain() {
	if r := recover(); r != nil {
		b.logger.Warn("recovering brain main goroutine")
		go b.brainMain()
	}
}

func (b *Brain) brainMain() {
	defer b.recoverBrainMain()

LOOP:
	for {
		select {
		case <-b.closeCh:
			b.ticker.Stop()
			break LOOP
		case <-b.ticker.C:
			unsortedInfo, err := b.gsMap.UnsortedInfo()
			if err != nil {
				b.logger.Panicf(
					"%s: this means stored type in map was not *GS",
					err.Error(),
				)
			}

			infos := unsortedInfo.Infos
			now := time.Now()
			for i, count := 0, len(infos); i < count; i++ {
				info := infos[i]
				shutdown := info.Fatal

				if !shutdown {
					if now.Sub(info.Summary.TimeLastCommunicate) >= b.params.MinimumWaitForClose {
						shutdown = true
					}
				}

				if !shutdown {
					if now.Sub(info.Summary.TimeStarted) >= b.params.MinimumWaitForClose {
						shutdown = info.Summary.TimeEstablished.IsZero() ||
							info.Summary.ConnectionCount == 0
					}
				}

				if shutdown {
					if err = b.Shutdown(info.ID); err != nil {
						b.logger.Panicf(
							"%s: this error means id was not found in map, the process will remain as zombie",
							err.Error(),
						)
					}
				}
			}
		}
	}

	b.logger.Info("brain closed")
}
