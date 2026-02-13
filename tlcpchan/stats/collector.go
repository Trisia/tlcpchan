package stats

import (
	"sync"
	"sync/atomic"
	"time"
)

type Snapshot struct {
	Timestamp         time.Time `json:"timestamp"`
	TotalConnections  int64     `json:"total_connections"`
	ActiveConnections int64     `json:"active_connections"`
	BytesReceived     int64     `json:"bytes_received"`
	BytesSent         int64     `json:"bytes_sent"`
	Requests          int64     `json:"requests"`
	Errors            int64     `json:"errors"`
	AvgLatency        int64     `json:"avg_latency_ns"`
	MaxLatency        int64     `json:"max_latency_ns"`
	MinLatency        int64     `json:"min_latency_ns"`
}

type Collector struct {
	enabled           atomic.Bool
	totalConnections  atomic.Int64
	activeConnections atomic.Int64
	bytesReceived     atomic.Int64
	bytesSent         atomic.Int64
	requests          atomic.Int64
	errors            atomic.Int64

	latencySum   atomic.Int64
	latencyCount atomic.Int64
	maxLatency   atomic.Int64
	minLatency   atomic.Int64

	snapshots    []Snapshot
	snapshotsMu  sync.RWMutex
	maxSnapshots int

	stopChan chan struct{}
	wg       sync.WaitGroup
	mu       sync.Mutex
}

var (
	defaultCollector *Collector
	once             sync.Once
)

func DefaultCollector() *Collector {
	once.Do(func() {
		defaultCollector = NewCollector(1000)
	})
	return defaultCollector
}

func NewCollector(maxSnapshots int) *Collector {
	c := &Collector{
		snapshots:    make([]Snapshot, 0, maxSnapshots),
		maxSnapshots: maxSnapshots,
		stopChan:     make(chan struct{}),
	}
	c.enabled.Store(true)
	c.minLatency.Store(-1)
	return c
}

func (c *Collector) Enable() {
	c.enabled.Store(true)
}

func (c *Collector) Disable() {
	c.enabled.Store(false)
}

func (c *Collector) IsEnabled() bool {
	return c.enabled.Load()
}

func (c *Collector) IncrementConnections() {
	if !c.enabled.Load() {
		return
	}
	c.totalConnections.Add(1)
	c.activeConnections.Add(1)
}

func (c *Collector) DecrementConnections() {
	if !c.enabled.Load() {
		return
	}
	c.activeConnections.Add(-1)
}

func (c *Collector) AddBytesReceived(n int64) {
	if !c.enabled.Load() {
		return
	}
	c.bytesReceived.Add(n)
}

func (c *Collector) AddBytesSent(n int64) {
	if !c.enabled.Load() {
		return
	}
	c.bytesSent.Add(n)
}

func (c *Collector) IncrementRequests() {
	if !c.enabled.Load() {
		return
	}
	c.requests.Add(1)
}

func (c *Collector) IncrementErrors() {
	if !c.enabled.Load() {
		return
	}
	c.errors.Add(1)
}

func (c *Collector) RecordLatency(latency time.Duration) {
	if !c.enabled.Load() {
		return
	}
	ns := latency.Nanoseconds()
	c.latencySum.Add(ns)
	c.latencyCount.Add(1)

	for {
		current := c.maxLatency.Load()
		if ns <= current {
			break
		}
		if c.maxLatency.CompareAndSwap(current, ns) {
			break
		}
	}

	for {
		current := c.minLatency.Load()
		if current != -1 && ns >= current {
			break
		}
		if c.minLatency.CompareAndSwap(current, ns) {
			break
		}
	}
}

func (c *Collector) GetSnapshot() Snapshot {
	var avgLatency int64
	count := c.latencyCount.Load()
	if count > 0 {
		avgLatency = c.latencySum.Load() / count
	}

	minLat := c.minLatency.Load()
	if minLat == -1 {
		minLat = 0
	}

	return Snapshot{
		Timestamp:         time.Now(),
		TotalConnections:  c.totalConnections.Load(),
		ActiveConnections: c.activeConnections.Load(),
		BytesReceived:     c.bytesReceived.Load(),
		BytesSent:         c.bytesSent.Load(),
		Requests:          c.requests.Load(),
		Errors:            c.errors.Load(),
		AvgLatency:        avgLatency,
		MaxLatency:        c.maxLatency.Load(),
		MinLatency:        minLat,
	}
}

func (c *Collector) StartSnapshotScheduler(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.stopChan:
		c.stopChan = make(chan struct{})
	default:
		return
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.saveSnapshot()
			case <-c.stopChan:
				return
			}
		}
	}()
}

func (c *Collector) StopSnapshotScheduler() {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.stopChan:
		return
	default:
		close(c.stopChan)
	}
	c.wg.Wait()
}

func (c *Collector) saveSnapshot() {
	snapshot := c.GetSnapshot()

	c.snapshotsMu.Lock()
	defer c.snapshotsMu.Unlock()

	if len(c.snapshots) >= c.maxSnapshots {
		c.snapshots = c.snapshots[1:]
	}
	c.snapshots = append(c.snapshots, snapshot)
}

func (c *Collector) GetSnapshots() []Snapshot {
	c.snapshotsMu.RLock()
	defer c.snapshotsMu.RUnlock()

	result := make([]Snapshot, len(c.snapshots))
	copy(result, c.snapshots)
	return result
}

func (c *Collector) GetSnapshotsCount() int {
	c.snapshotsMu.RLock()
	defer c.snapshotsMu.RUnlock()
	return len(c.snapshots)
}

func (c *Collector) ClearSnapshots() {
	c.snapshotsMu.Lock()
	defer c.snapshotsMu.Unlock()
	c.snapshots = make([]Snapshot, 0, c.maxSnapshots)
}

func (c *Collector) Reset() {
	c.totalConnections.Store(0)
	c.activeConnections.Store(0)
	c.bytesReceived.Store(0)
	c.bytesSent.Store(0)
	c.requests.Store(0)
	c.errors.Store(0)
	c.latencySum.Store(0)
	c.latencyCount.Store(0)
	c.maxLatency.Store(0)
	c.minLatency.Store(-1)
	c.ClearSnapshots()
}

type Stats struct {
	TotalConnections  int64   `json:"total_connections"`
	ActiveConnections int64   `json:"active_connections"`
	BytesReceived     int64   `json:"bytes_received"`
	BytesSent         int64   `json:"bytes_sent"`
	Requests          int64   `json:"requests"`
	Errors            int64   `json:"errors"`
	AvgLatencyMs      float64 `json:"avg_latency_ms"`
	MaxLatencyMs      float64 `json:"max_latency_ms"`
	MinLatencyMs      float64 `json:"min_latency_ms"`
}

func (c *Collector) GetStats() Stats {
	snapshot := c.GetSnapshot()
	return Stats{
		TotalConnections:  snapshot.TotalConnections,
		ActiveConnections: snapshot.ActiveConnections,
		BytesReceived:     snapshot.BytesReceived,
		BytesSent:         snapshot.BytesSent,
		Requests:          snapshot.Requests,
		Errors:            snapshot.Errors,
		AvgLatencyMs:      float64(snapshot.AvgLatency) / 1e6,
		MaxLatencyMs:      float64(snapshot.MaxLatency) / 1e6,
		MinLatencyMs:      float64(snapshot.MinLatency) / 1e6,
	}
}

func Enable()                       { DefaultCollector().Enable() }
func Disable()                      { DefaultCollector().Disable() }
func IsEnabled() bool               { return DefaultCollector().IsEnabled() }
func IncrementConnections()         { DefaultCollector().IncrementConnections() }
func DecrementConnections()         { DefaultCollector().DecrementConnections() }
func AddBytesReceived(n int64)      { DefaultCollector().AddBytesReceived(n) }
func AddBytesSent(n int64)          { DefaultCollector().AddBytesSent(n) }
func IncrementRequests()            { DefaultCollector().IncrementRequests() }
func IncrementErrors()              { DefaultCollector().IncrementErrors() }
func RecordLatency(d time.Duration) { DefaultCollector().RecordLatency(d) }
func GetStats() Stats               { return DefaultCollector().GetStats() }
func GetSnapshot() Snapshot         { return DefaultCollector().GetSnapshot() }
func GetSnapshots() []Snapshot      { return DefaultCollector().GetSnapshots() }
