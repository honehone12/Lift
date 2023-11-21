package monitor

type MonitoringMessage struct {
	GuidRaw            []byte
	ConnectionCount    int
	SessionCount       int
	ActiveSessionCount int
}
