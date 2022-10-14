package model

import "time"

type Metrics struct {
	LastUpdate int64 `json:"last_update"`
	Sessions   map[string]SessionMetrics
}

func (m *Metrics) PushSession(svcName string, sm SessionMetrics) {
	if m.Sessions == nil {
		m.Sessions = make(map[string]SessionMetrics)
	}
	if prev, found := m.Sessions[svcName]; found {
		prev.BytesRead += sm.BytesRead
		prev.BytesWritten += sm.BytesWritten
		prev.LastUpdate = sm.LastUpdate
		m.Sessions[svcName] = prev
	} else {
		m.Sessions[svcName] = sm
	}
	m.LastUpdate = time.Now().UnixMilli()
}

type SessionMetrics struct {
	BytesRead    int64
	BytesWritten int64
	LastUpdate   int64
}
