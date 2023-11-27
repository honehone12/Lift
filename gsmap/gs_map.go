package gsmap

import (
	"errors"
	"lift/gsmap/gs"
	"lift/gsmap/gsparams"
	"lift/logger"
	"sync"
)

type GSMap struct {
	count  int
	inner  *sync.Map
	logger logger.Logger
}

var (
	ErrorNoSuchItem = errors.New("no such item")
	ErrorCastFail   = errors.New("failed to cast item")
)

func NewGSMap(logger logger.Logger) *GSMap {
	return &GSMap{
		count:  0,
		inner:  &sync.Map{},
		logger: logger,
	}
}

func (m *GSMap) Count() int {
	return m.count
}

func (m *GSMap) Launch(params *gsparams.GSParams) error {
	gs, err := gs.NewGS(params, m.logger)
	if err != nil {
		return err
	}

	if gs.StartProcess(); err != nil {
		return err
	}

	m.add(params.UuidString(), gs)

	return nil
}

func (m *GSMap) add(id string, gs *gs.GS) {
	if _, exists := m.inner.LoadOrStore(id, gs); !exists {
		m.count++
	}
}

func (m *GSMap) remove(id string) {
	if _, exists := m.inner.LoadAndDelete(id); exists {
		m.count--
	}
}

func (m *GSMap) Item(id string) (*gs.GS, error) {
	i, ok := m.inner.Load(id)
	if ok {
		gs, ok := i.(*gs.GS)
		if ok {
			return gs, nil
		} else {
			return nil, ErrorCastFail
		}

	} else {
		return nil, ErrorNoSuchItem
	}
}
