package gsparams

import libuuid "github.com/google/uuid"

type GSParams struct {
	Process string
	Uuid    libuuid.UUID
	Address string
	Port    string
}

func NewGSParams(
	process string,
	uuid libuuid.UUID,
	address string,
	port string,
) *GSParams {
	return &GSParams{
		Process: process,
		Uuid:    uuid,
		Address: address,
		Port:    port,
	}
}

func (p *GSParams) ProcessName() string {
	return p.Process
}

func (p *GSParams) UuidString() string {
	return p.Uuid.String()
}

func (p *GSParams) ToArgs() []string {
	return []string{
		"-a", p.Address,
		"-p", p.Port,
		"-u", p.Uuid.String(),
	}
}
