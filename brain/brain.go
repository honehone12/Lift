package brain

import (
	"errors"
	"lift/brain/portman"
	"lift/gsmap"
	"lift/gsmap/gs"
	"lift/gsmap/gsinfo"
	"lift/gsmap/gsparams"
	"lift/logger"
	"lift/setting"
	"sort"
	"time"

	libuuid "github.com/google/uuid"
)

type BrainParams struct {
	GSProcess        []setting.GSSetting
	GSListenAddress  string
	GSMessageTimeout time.Duration

	PortParams          portman.PortManParams
	LoopInterval        time.Duration
	MinimumWaitForClose time.Duration
}

type Brain struct {
	params  *BrainParams
	portMan *portman.PortMan

	gsMap   *gsmap.GSMap
	logger  logger.Logger
	ticker  *time.Ticker
	closeCh chan bool
}

var (
	ErrorIndexOutOfRange = errors.New("index is out of range GSProcess")
)

func GenerateId() [16]byte {
	return libuuid.New()
}

func NewBrain(
	params *BrainParams,
	gsMap *gsmap.GSMap,
	logger logger.Logger,
) (*Brain, error) {
	pm, err := portman.NewPortMan(params.PortParams)
	if err != nil {
		return nil, err
	}

	b := &Brain{
		params:  params,
		portMan: pm,
		gsMap:   gsMap,
		logger:  logger,
		ticker:  time.NewTicker(params.LoopInterval),
		closeCh: make(chan bool),
	}

	go b.brainMain()

	return b, nil
}

func (b *Brain) PortMan() *portman.PortMan {
	return b.portMan
}

func (b *Brain) ValidIndex(idx int) bool {
	return idx >= 0 && idx < len(b.params.GSProcess)
}

func (b *Brain) Launch(idx int) (*gsinfo.GSPort, error) {
	if idx < 0 || idx >= len(b.params.GSProcess) {
		return nil, ErrorIndexOutOfRange
	}

	p, err := b.portMan.Next()
	if err != nil {
		return nil, err
	}

	param := gsparams.NewGSParams(
		idx,
		b.params.GSProcess[idx].ProcessName,
		GenerateId(),
		b.params.GSListenAddress,
		p,
		b.params.GSMessageTimeout,
	)
	gs, err := gs.NewGS(param, b.logger)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	b.gsMap.Add(id, gs)
	return &gsinfo.GSPort{
		Id:   id,
		Port: p.Number(),
	}, nil
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
					if err = b.Shutdown(info.Id); err != nil {
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

func (b *Brain) BackfillList(idx int) ([]gsinfo.GSBackfillPort, error) {
	if idx < 0 || idx >= len(b.params.GSProcess) {
		return nil, ErrorIndexOutOfRange
	}

	unsortedInfo, err := b.gsMap.UnsortedInfo()
	if err != nil {
		return nil, err
	}

	total := len(unsortedInfo.Infos)
	temp := make([]int, 0, total)
	cap := b.params.GSProcess[idx].ConnectionCapacity
	for i := 0; i < total; i++ {
		info := unsortedInfo.Infos[i]
		if info.Index != idx {
			continue
		}
		if cap-info.Summary.ConnectionCount <= 0 {
			continue
		}

		temp = append(temp, i)
	}

	sort.SliceStable(temp, func(i, j int) bool {
		infoI := unsortedInfo.Infos[temp[i]]
		infoJ := unsortedInfo.Infos[temp[j]]
		roomI := cap - infoI.Summary.ConnectionCount
		roomJ := cap - infoJ.Summary.ConnectionCount

		if roomI == roomJ {
			return infoI.Summary.TimeStarted.Before(infoJ.Summary.TimeStarted)
		} else {
			return roomI < roomJ
		}
	})

	count := len(temp)
	buff := make([]gsinfo.GSBackfillPort, 0, count)
	for i := 0; i < count; i++ {
		info := unsortedInfo.Infos[temp[i]]
		buff = append(buff, gsinfo.GSBackfillPort{
			GsPort: gsinfo.GSPort{
				Id:   info.Id,
				Port: info.Port,
			},
			Since:  info.Summary.TimeStarted,
			Active: info.Summary.ActiveSessionCount,
		})
	}

	return buff, nil
}
