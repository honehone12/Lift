package gsmap

import (
	"errors"
	"lift/gsmap/gs"
	"sync"
)

type GSMap struct {
	count int
	inner *sync.Map
}

var (
	ErrorNoSuchItem = errors.New("no such item")
	ErrorCastFail   = errors.New("failed to cast item")
)

func NewGSMap() *GSMap {
	return &GSMap{
		count: 0,
		inner: &sync.Map{},
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
