package stats

import (
	"sync"
	"sync/atomic"
	"time"
)

// Snapshot 统计快照，记录某一时刻的统计信息
type Snapshot struct {
	// Timestamp 快照时间
	Timestamp time.Time `json:"timestamp"`
	// TotalConnections 累计连接总数
	TotalConnections int64 `json:"totalConnections"`
	// ActiveConnections 当前活跃连接数
	ActiveConnections int64 `json:"activeConnections"`
	// BytesReceived 累计接收字节数，单位: 字节
	BytesReceived int64 `json:"bytesReceived"`
	// BytesSent 累计发送字节数，单位: 字节
	BytesSent int64 `json:"bytesSent"`
	// Requests 累计请求数（HTTP代理）
	Requests int64 `json:"requests"`
	// Errors 累计错误数
	Errors int64 `json:"errors"`
	// AvgLatency 平均延迟，单位: 纳秒
	AvgLatency int64 `json:"avgLatencyNs"`
	// MaxLatency 最大延迟，单位: 纳秒
	MaxLatency int64 `json:"maxLatencyNs"`
	// MinLatency 最小延迟，单位: 纳秒
	MinLatency int64 `json:"minLatencyNs"`
}

// Collector 统计信息收集器
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

// DefaultCollector 获取默认统计收集器单例
// 返回:
//   - *Collector: 默认统计收集器实例
func DefaultCollector() *Collector {
	once.Do(func() {
		defaultCollector = NewCollector(1000)
	})
	return defaultCollector
}

// NewCollector 创建新的统计收集器
// 参数:
//   - maxSnapshots: 最大快照保留数量
//
// 返回:
//   - *Collector: 统计收集器实例
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

// RecordLatency 记录延迟数据
// 参数:
//   - latency: 延迟时间
//
// 注意: 会自动更新最大/最小延迟
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

// GetSnapshot 获取当前统计快照
// 返回:
//   - Snapshot: 当前统计信息快照
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

// StartSnapshotScheduler 启动快照定时调度器
// 参数:
//   - interval: 快照间隔时间
//
// 注意: 该方法启动后台goroutine，调用StopSnapshotScheduler()停止
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

// StopSnapshotScheduler 停止快照定时调度器
// 注意: 该方法会等待后台goroutine退出后返回
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

// Stats 统计信息DTO，用于API返回
type Stats struct {
	// TotalConnections 累计连接总数
	TotalConnections int64 `json:"totalConnections"`
	// ActiveConnections 当前活跃连接数
	ActiveConnections int64 `json:"activeConnections"`
	// BytesReceived 累计接收字节数，单位: 字节
	BytesReceived int64 `json:"bytesReceived"`
	// BytesSent 累计发送字节数，单位: 字节
	BytesSent int64 `json:"bytesSent"`
	// Requests 累计请求数（HTTP代理）
	Requests int64 `json:"requests"`
	// Errors 累计错误数
	Errors int64 `json:"errors"`
	// AvgLatencyMs 平均延迟，单位: 毫秒
	AvgLatencyMs float64 `json:"avgLatencyMs"`
	// MaxLatencyMs 最大延迟，单位: 毫秒
	MaxLatencyMs float64 `json:"maxLatencyMs"`
	// MinLatencyMs 最小延迟，单位: 毫秒
	MinLatencyMs float64 `json:"minLatencyMs"`
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
