package gsinfo

import "time"

type MonitoringSummary struct {
	TimeStarted         time.Time
	TimeEstablished     time.Time
	TimeLastCommunicate time.Time

	ConnectionCount    int64
	SessionCount       int64
	ActiveSessionCount int64
}

type GSInfo struct {
	ID      string
	Port    uint16
	Summary MonitoringSummary
	Fatal   bool
}

type AllGSInfo struct {
	Count int64

	Infos []GSInfo
}
