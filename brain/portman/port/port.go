package port

import "strconv"

type Port struct {
	number uint16
}

func NilPort() Port {
	return Port{
		number: 0,
	}
}

func NewPort(n uint16) Port {
	return Port{
		number: n,
	}
}

func (p Port) Number() uint16 {
	return p.number
}

func (p Port) String() string {
	return strconv.FormatUint(uint64(p.number), 10)
}

func (p Port) Empty() bool {
	return p.number == 0
}
