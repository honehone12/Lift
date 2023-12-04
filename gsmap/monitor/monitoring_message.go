package monitor

const (
	ErrorFatal uint8 = iota
	ErrorWarn
	NoError
)

type MonitoringMessage struct {
	GuidRaw            []byte
	ConnectionCount    int64
	SessionCount       int64
	ActiveSessionCount int64

	ErrorCode uint8
	ErrorUtf8 []byte
}
