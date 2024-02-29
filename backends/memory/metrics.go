package memory

import (
	"fmt"
	"runtime"
	"sync/atomic"
)

// Metrics keeps track of counts of things
type Metrics struct {
	// This is across all subsystems
	Vertices int   `json:"vertices"`
	Writes   int64 `json:"writes"`
	Reads    int64 `json:"reads"`

	// Resource specific metrics
	ResourceSummary map[string]Summary
}

// Debugging function to print stats
// example usage:
// var mem runtime.MemStats
// printMemoryStats(mem)
// runtime.GC()
//
//lint:ignore U1000 Ignore unused function temporarily for debugging
func printMemoryStats(mem runtime.MemStats) {
	runtime.ReadMemStats(&mem)
	fmt.Printf("mem.Alloc: %d\n", mem.Alloc)
	fmt.Printf("mem.TotalAlloc (cumulative): %d\n", mem.TotalAlloc)
	fmt.Printf("mem.HeapAlloc: %d\n", mem.HeapAlloc)
	fmt.Printf("mem.NumGC: %d\n\n", mem.NumGC)
}

// Resource summary to hold counts for each type
// We assemble this as we create a new graph
type Summary struct {
	Name   string
	Counts map[string]int64
}

// NewResource resets the resource counters for a cluster
func (m *Metrics) NewResource(cluster string) {
	m.ResourceSummary[cluster] = Summary{Name: cluster}
}

func (m *Metrics) AddResource() {

}

func (m *Metrics) IncWriteCount() *Metrics {
	atomic.AddInt64(&m.Writes, 1)
	return m
}
func (m *Metrics) IncReadCount() *Metrics {
	atomic.AddInt64(&m.Reads, 1)
	return m
}
