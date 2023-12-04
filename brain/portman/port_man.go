package portman

import (
	"errors"
	"sync"

	"lift/brain/portman/port"

	libqueue "github.com/Workiva/go-datastructures/queue"
)

type PortManParams struct {
	InitialCapacity int64
	StartFrom       uint16
}

type PortMan struct {
	params PortManParams
	q      *libqueue.Queue
	m      *sync.Map
}

type PortInfo struct {
	CurrentCapacity int64
	Peek            uint16
}

var (
	ErrorInvalidPortZero = errors.New("invalid port number zero")
	ErrorCastFail        = errors.New("failed to cast interface")
	ErrorPortInUse       = errors.New("port is already in use")
	ErrorPortNotUsed     = errors.New("port is not used")
)

func NewPortMan(params PortManParams) (*PortMan, error) {
	if params.InitialCapacity == 0 {
		return nil, errors.New("zero capacity")
	}
	if params.StartFrom == 0 {
		return nil, ErrorInvalidPortZero
	}

	q := libqueue.New(params.InitialCapacity)
	m := &sync.Map{}

	pn := params.StartFrom
	var err error
	for i, count := int64(0), params.InitialCapacity; i < count; i++ {
		p := port.NewPort(pn)
		if err = q.Put(p); err != nil {
			return nil, err
		}
		m.Store(pn, false)
		pn++
	}

	return &PortMan{
		q:      q,
		m:      m,
		params: params,
	}, nil
}

func (pm *PortMan) Info() (PortInfo, error) {
	i, err := pm.q.Peek()
	if err != nil {
		return PortInfo{
			CurrentCapacity: 0,
			Peek:            0,
		}, nil
	}

	peek, ok := i.(port.Port)
	if !ok {
		return PortInfo{}, ErrorCastFail
	}

	return PortInfo{
		CurrentCapacity: pm.q.Len(),
		Peek:            peek.Number(),
	}, nil
}

func (pm *PortMan) Next() (port.Port, error) {
	i, err := pm.q.Get(1)
	if err != nil {
		return port.Port{}, err
	}

	p, ok := i[0].(port.Port)
	if !ok {
		return port.Port{}, ErrorCastFail
	}

	ok = pm.m.CompareAndSwap(p.Number(), false, true)
	if !ok {
		return port.Port{}, ErrorPortInUse
	}

	return p, nil
}

func (pm *PortMan) Return(p port.Port) error {
	if p.Empty() {
		return ErrorInvalidPortZero
	}

	ok := pm.m.CompareAndSwap(p.Number(), true, false)
	if !ok {
		return ErrorPortNotUsed
	}

	return pm.q.Put(p)
}
