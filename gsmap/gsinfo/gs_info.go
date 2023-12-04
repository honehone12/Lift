package gsinfo

import "time"

const (
	MonitoringStatusError uint8 = iota
	MonitoringStatusConnectionError
	MonitoringStatusNotConnected
	MonitoringStatusOK
	MonitoringStatusClosed
)

const (
	ProcessStatusError uint8 = iota
	ProcessStatusNotStarted
	ProcessStatusOK
	ProcessStatusCanceled
)

type MonitoringSummary struct {
	MonitoringStatus uint8

	TimeStarted     time.Time
	TimeEstablished time.Time
	TimeClosed      time.Time

	ConnectionCount    int64
	SessionCount       int64
	ActiveSessionCount int64
}

type GSInfo struct {
	ID            string
	Port          uint16
	ProcessStatus uint8
	Summary       MonitoringSummary
}

type AllGSInfo struct {
	Count int64

	Infos []GSInfo
}
