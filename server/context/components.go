package context

import (
	"lift/gsmap"

	"github.com/gorilla/websocket"
)

type Components struct {
	metadata   *Metadata
	wsUpgrader *websocket.Upgrader
	gsMap      *gsmap.GSMap
}

func NewComponents(m *Metadata) *Components {
	return &Components{
		metadata:   m,
		wsUpgrader: &websocket.Upgrader{},
		gsMap:      gsmap.NewGSMap(),
	}
}

func (c *Components) Metadata() *Metadata {
	return c.metadata
}

func (c *Components) WebSocketUpgrader() *websocket.Upgrader {
	return c.wsUpgrader
}

func (c *Components) GSMap() *gsmap.GSMap {
	return c.gsMap
}
