package setting

type GSSetting struct {
	ProcessName        string
	ConnectionCapacity int64
}

type Setting struct {
	LogLevel int

	ServiceName     string
	ServiceVersion  string
	ServiceListenAt string

	GSProcess           []GSSetting
	GSListenAddress     string
	GSMessageTimeoutSec int

	PortCapacity  int64
	PortStartFrom uint16

	BrainIntervalSec    int
	BrainMinimumWaitSec int
}
