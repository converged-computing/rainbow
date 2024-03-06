package memory

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sync/atomic"
)

// CountResource under some named top level, usually a root-> cluster (the cluster)
func (m *Metrics) CountResource(topLevel, resourceType string) {
	summary, ok := m.ResourceSummary[topLevel]
	if !ok {
		summary = Summary{Name: topLevel, Counts: map[string]int64{}}
	}

	// Here we are updating the count for the specific type
	count, ok := summary.Counts[resourceType]
	if !ok {
		count = 0
	}
	count += 1
	summary.Counts[resourceType] = count
	m.ResourceSummary[topLevel] = summary
}

// NewResource resets the resource counters for a cluster
func (m *Metrics) NewResource(levelName string) {
	counts := map[string]int64{}
	m.ResourceSummary[levelName] = Summary{Name: levelName, Counts: counts}
}

// Show prints a summary of resources for an entire subsystem
func (m *Metrics) Show() error {
	out, err := json.MarshalIndent(m.ResourceSummary, "", " ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func (m *Metrics) IncWriteCount() *Metrics {
	atomic.AddInt64(&m.Writes, 1)
	return m
}
func (m *Metrics) IncReadCount() *Metrics {
	atomic.AddInt64(&m.Reads, 1)
	return m
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
