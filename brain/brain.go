package brain

import (
	"lift/brain/portman"
	"lift/gsmap"
	"lift/gsmap/gsparams"
	"time"

	libuuid "github.com/google/uuid"
)

type BrainParams struct {
	GSProcessName       string
	GSListenAddress     string
	GSMonitoringTimeout time.Duration

	PortParams portman.PortManParams
}

type Brain struct {
	params  *BrainParams
	portMan *portman.PortMan

	gsMap *gsmap.GSMap
}

func GenerateId() [16]byte {
	return libuuid.New()
}

func NewBrain(params *BrainParams, gsMap *gsmap.GSMap) (*Brain, error) {
	pm, err := portman.NewPortMan(params.PortParams)
	if err != nil {
		return nil, err
	}

	return &Brain{
		params:  params,
		portMan: pm,
		gsMap:   gsMap,
	}, nil
}

func (b *Brain) PortMan() *portman.PortMan {
	return b.portMan
}

func (b *Brain) Launch(n int) error {
	for i := 0; i < n; i++ {
		p, err := b.portMan.Next()
		if err != nil {
			return err
		}

		if err = b.gsMap.Launch(gsparams.NewGSParams(
			b.params.GSProcessName,
			GenerateId(),
			b.params.GSListenAddress,
			p,
			b.params.GSMonitoringTimeout,
		)); err != nil {
			return err
		}
	}
	return nil
}
