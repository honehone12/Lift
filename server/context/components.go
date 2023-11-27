package context

import (
	"lift/brain"
	"lift/gsmap"

	"github.com/gorilla/websocket"
)

type Components struct {
	metadata   *Metadata
	wsUpgrader *websocket.Upgrader
	gsMap      *gsmap.GSMap
	brain      *brain.Brain
}

func NewComponents(
	m *Metadata,
	gsm *gsmap.GSMap,
	b *brain.Brain,
) *Components {
	return &Components{
		metadata:   m,
		wsUpgrader: &websocket.Upgrader{},
		gsMap:      gsm,
		brain:      b,
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

func (c *Components) Brain() *brain.Brain {
	return c.brain
}
