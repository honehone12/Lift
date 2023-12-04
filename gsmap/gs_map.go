package gsmap

import (
	"errors"
	"lift/gsmap/gs"
	"lift/gsmap/gsinfo"
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

func (m *GSMap) Add(id string, gs *gs.GS) {
	if _, exists := m.inner.LoadOrStore(id, gs); !exists {
		m.count++
	}
}

func (m *GSMap) Remove(id string) {
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

func (m *GSMap) UnsortedInfo() (*gsinfo.AllGSInfo, error) {
	info := &gsinfo.AllGSInfo{
		Count: int64(m.count),
		Infos: make([]gsinfo.GSInfo, 0, m.count),
	}

	var err error
	m.inner.Range(func(k interface{}, v interface{}) bool {
		gs, ok := v.(*gs.GS)
		if !ok {
			err = ErrorCastFail
			return false
		}
		info.Infos = append(info.Infos, gs.Info())
		return true
	})
	if err != nil {
		return nil, err
	}

	return info, nil
}
