package setting

type GSExecutable struct {
	ProcessName        string
	ConnectionCapacity int64
	MaxBackfillSec     int
}

type Setting struct {
	LogLevel int

	ServiceName     string
	ServiceVersion  string
	ServiceListenAt string

	GSExecutables       []GSExecutable
	GSListenAddress     string
	GSMessageTimeoutSec int

	PortCapacity  int64
	PortStartFrom uint16

	BrainIntervalSec    int
	BrainMinimumWaitSec int
}
