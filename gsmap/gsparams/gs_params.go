package gsparams

import (
	"fmt"
	"time"

	libuuid "github.com/google/uuid"
)

type GSParams struct {
	process string
	uuid    libuuid.UUID
	address string
	port    string

	monitoringTimeout time.Duration
}

func NewGSParams(
	process string,
	uuid libuuid.UUID,
	address string,
	port string,
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
	return p.uuid.String()
}

func (p *GSParams) ToArgs() []string {
	return []string{
		"-a", p.address,
		"-p", p.port,
		"-u", p.uuid.String(),
	}
}

func (p *GSParams) NextMonitoringTimeout() time.Time {
	return time.Now().Add(p.monitoringTimeout)
}

func (p *GSParams) LogFormat(msg string) string {
	return fmt.Sprintf("GS PROCESS [%s] ", p.uuid.String()) + msg
}
