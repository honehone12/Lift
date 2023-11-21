package gs

import (
	"lift/gs/gsparams"
	"lift/gs/gsprocess"

	"github.com/gorilla/websocket"
)

type GS struct {
	params  *gsparams.GSParams
	process *gsprocess.GSProcess
	conn    *websocket.Conn
}

func NewGS(
	params *gsparams.GSParams,
	process *gsprocess.GSProcess,
) *GS {
	return &GS{
		params:  params,
		process: process,
	}
}

func (gs *GS) Established() bool {
	return gs.conn != nil
}

func (gs *GS) SetWSConnection(conn *websocket.Conn) {
	if conn != nil && gs.conn == nil {
		gs.conn = conn
	}
	if conn == nil && gs.conn != nil {
		gs.conn = nil
	}
}
