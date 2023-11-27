package gsparams

import (
	"fmt"
	"lift/brain/portman/port"
	"time"

	libuuid "github.com/google/uuid"
)

type GSParams struct {
	process string
	uuid    [16]byte
	address string
	port    port.Port

	monitoringTimeout time.Duration
}

func NewGSParams(
	process string,
	uuid [16]byte,
	address string,
	port port.Port,
	monitoringTimeout time.Duration,
) *GSParams {
	return &GSParams{
		process:           process,
		uuid:              uuid,
		address:           address,
		port:              port,
		monitoringTimeout: monitoringTimeout,
	}
}

func (p *GSParams) ProcessName() string {
	return p.process
}

func (p *GSParams) UuidString() string {
	return libuuid.UUID(p.uuid).String()
}

func (p *GSParams) ToArgs() []string {
	return []string{
		"-a", p.address,
		"-p", p.port.String(),
		"-u", p.UuidString(),
	}
}

func (p *GSParams) NextMonitoringTimeout() time.Time {
	return time.Now().Add(p.monitoringTimeout)
}

func (p *GSParams) LogFormat(msg string) string {
	return fmt.Sprintf("GS PROCESS [%s] ", p.UuidString()) + msg
}
