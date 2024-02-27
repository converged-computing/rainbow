package memory

import (
	"sync/atomic"
)

// Metrics keeps track of counts of things
type Metrics struct {
	Vertices int   `json:"vertices"`
	Writes   int64 `json:"writes"`
	Reads    int64 `json:"reads"`
}

func (m *Metrics) IncWriteCount() *Metrics {
	atomic.AddInt64(&m.Writes, 1)
	return m
}
func (m *Metrics) IncReadCount() *Metrics {
	atomic.AddInt64(&m.Reads, 1)
	return m
}
