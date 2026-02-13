package stats

import (
	"sync"
	"testing"
	"time"
)

func TestNewCollector(t *testing.T) {
	c := NewCollector(100)
	if c == nil {
		t.Fatal("NewCollector() 返回 nil")
	}

	if !c.IsEnabled() {
		t.Error("新建的 Collector 应默认启用")
	}
}

func TestCollectorEnableDisable(t *testing.T) {
	c := NewCollector(100)

	c.Disable()
	if c.IsEnabled() {
		t.Error("禁用后 IsEnabled() 应返回 false")
	}

	c.Enable()
	if !c.IsEnabled() {
		t.Error("启用后 IsEnabled() 应返回 true")
	}
}

func TestIncrementConnections(t *testing.T) {
	c := NewCollector(100)
	c.Reset()

	c.IncrementConnections()
	snapshot := c.GetSnapshot()

	if snapshot.TotalConnections != 1 {
		t.Errorf("TotalConnections 应为 1, 实际为 %d", snapshot.TotalConnections)
	}

	if snapshot.ActiveConnections != 1 {
		t.Errorf("ActiveConnections 应为 1, 实际为 %d", snapshot.ActiveConnections)
	}
}

func TestDecrementConnections(t *testing.T) {
	c := NewCollector(100)
	c.Reset()

	c.IncrementConnections()
	c.IncrementConnections()
	c.DecrementConnections()
	snapshot := c.GetSnapshot()

	if snapshot.TotalConnections != 2 {
		t.Errorf("TotalConnections 应为 2, 实际为 %d", snapshot.TotalConnections)
	}

	if snapshot.ActiveConnections != 1 {
		t.Errorf("ActiveConnections 应为 1, 实际为 %d", snapshot.ActiveConnections)
	}
}

func TestBytesCounters(t *testing.T) {
	c := NewCollector(100)
	c.Reset()

	c.AddBytesReceived(100)
	c.AddBytesReceived(200)
	c.AddBytesSent(150)

	snapshot := c.GetSnapshot()

	if snapshot.BytesReceived != 300 {
		t.Errorf("BytesReceived 应为 300, 实际为 %d", snapshot.BytesReceived)
	}

	if snapshot.BytesSent != 150 {
		t.Errorf("BytesSent 应为 150, 实际为 %d", snapshot.BytesSent)
	}
}

func TestRequestsAndErrors(t *testing.T) {
	c := NewCollector(100)
	c.Reset()

	c.IncrementRequests()
	c.IncrementRequests()
	c.IncrementRequests()
	c.IncrementErrors()

	snapshot := c.GetSnapshot()

	if snapshot.Requests != 3 {
		t.Errorf("Requests 应为 3, 实际为 %d", snapshot.Requests)
	}

	if snapshot.Errors != 1 {
		t.Errorf("Errors 应为 1, 实际为 %d", snapshot.Errors)
	}
}

func TestRecordLatency(t *testing.T) {
	c := NewCollector(100)
	c.Reset()

	c.RecordLatency(100 * time.Millisecond)
	c.RecordLatency(200 * time.Millisecond)
	c.RecordLatency(50 * time.Millisecond)

	snapshot := c.GetSnapshot()

	if snapshot.AvgLatency != 116666666 {
		t.Errorf("AvgLatency 应为约 116666666ns, 实际为 %d", snapshot.AvgLatency)
	}

	if snapshot.MaxLatency != 200000000 {
		t.Errorf("MaxLatency 应为 200000000ns, 实际为 %d", snapshot.MaxLatency)
	}

	if snapshot.MinLatency != 50000000 {
		t.Errorf("MinLatency 应为 50000000ns, 实际为 %d", snapshot.MinLatency)
	}
}

func TestLatencyConcurrency(t *testing.T) {
	c := NewCollector(100)
	c.Reset()

	var wg sync.WaitGroup
	numGoroutines := 100
	latencyPerGoroutine := 1000

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < latencyPerGoroutine; j++ {
				c.RecordLatency(time.Duration(j+1) * time.Microsecond)
			}
		}()
	}

	wg.Wait()

	snapshot := c.GetSnapshot()
	expectedCount := int64(numGoroutines * latencyPerGoroutine)

	if c.latencyCount.Load() != expectedCount {
		t.Errorf("latencyCount 应为 %d, 实际为 %d", expectedCount, c.latencyCount.Load())
	}

	if snapshot.MaxLatency < 1 {
		t.Error("MaxLatency 应大于 0")
	}
}

func TestDisabledCollector(t *testing.T) {
	c := NewCollector(100)
	c.Reset()
	c.Disable()

	c.IncrementConnections()
	c.AddBytesReceived(100)
	c.AddBytesSent(100)
	c.IncrementRequests()
	c.IncrementErrors()
	c.RecordLatency(100 * time.Millisecond)

	snapshot := c.GetSnapshot()

	if snapshot.TotalConnections != 0 {
		t.Errorf("禁用状态下 TotalConnections 应为 0, 实际为 %d", snapshot.TotalConnections)
	}

	if snapshot.BytesReceived != 0 {
		t.Errorf("禁用状态下 BytesReceived 应为 0, 实际为 %d", snapshot.BytesReceived)
	}

	if snapshot.Requests != 0 {
		t.Errorf("禁用状态下 Requests 应为 0, 实际为 %d", snapshot.Requests)
	}
}

func TestReset(t *testing.T) {
	c := NewCollector(100)

	c.IncrementConnections()
	c.AddBytesReceived(1000)
	c.AddBytesSent(500)
	c.IncrementRequests()
	c.IncrementErrors()
	c.RecordLatency(100 * time.Millisecond)

	c.Reset()
	snapshot := c.GetSnapshot()

	if snapshot.TotalConnections != 0 {
		t.Errorf("重置后 TotalConnections 应为 0, 实际为 %d", snapshot.TotalConnections)
	}

	if snapshot.ActiveConnections != 0 {
		t.Errorf("重置后 ActiveConnections 应为 0, 实际为 %d", snapshot.ActiveConnections)
	}

	if snapshot.BytesReceived != 0 {
		t.Errorf("重置后 BytesReceived 应为 0, 实际为 %d", snapshot.BytesReceived)
	}

	if snapshot.BytesSent != 0 {
		t.Errorf("重置后 BytesSent 应为 0, 实际为 %d", snapshot.BytesSent)
	}

	if snapshot.Requests != 0 {
		t.Errorf("重置后 Requests 应为 0, 实际为 %d", snapshot.Requests)
	}

	if snapshot.Errors != 0 {
		t.Errorf("重置后 Errors 应为 0, 实际为 %d", snapshot.Errors)
	}
}

func TestGetSnapshot(t *testing.T) {
	c := NewCollector(100)
	c.Reset()

	c.IncrementConnections()
	c.DecrementConnections()
	c.AddBytesReceived(100)
	c.AddBytesSent(50)
	c.IncrementRequests()
	c.IncrementErrors()
	c.RecordLatency(10 * time.Millisecond)

	snapshot := c.GetSnapshot()

	if snapshot.Timestamp.IsZero() {
		t.Error("快照时间戳不应为零值")
	}

	if snapshot.TotalConnections != 1 {
		t.Errorf("TotalConnections 应为 1, 实际为 %d", snapshot.TotalConnections)
	}

	if snapshot.ActiveConnections != 0 {
		t.Errorf("ActiveConnections 应为 0, 实际为 %d", snapshot.ActiveConnections)
	}

	if snapshot.BytesReceived != 100 {
		t.Errorf("BytesReceived 应为 100, 实际为 %d", snapshot.BytesReceived)
	}

	if snapshot.BytesSent != 50 {
		t.Errorf("BytesSent 应为 50, 实际为 %d", snapshot.BytesSent)
	}
}

func TestGetStats(t *testing.T) {
	c := NewCollector(100)
	c.Reset()

	c.RecordLatency(1 * time.Millisecond)

	stats := c.GetStats()

	if stats.AvgLatencyMs != 1.0 {
		t.Errorf("AvgLatencyMs 应为 1.0, 实际为 %f", stats.AvgLatencyMs)
	}
}

func TestSnapshots(t *testing.T) {
	c := NewCollector(10)
	c.ClearSnapshots()

	c.StopSnapshotScheduler()
	c.StartSnapshotScheduler(5 * time.Millisecond)
	time.Sleep(80 * time.Millisecond)
	c.StopSnapshotScheduler()

	snapshots := c.GetSnapshots()

	if len(snapshots) < 1 {
		t.Errorf("应有至少1个快照, 实际有 %d 个", len(snapshots))
	}

	count := c.GetSnapshotsCount()
	if count < 1 {
		t.Errorf("GetSnapshotsCount 应返回至少1, 实际返回 %d", count)
	}
}

func TestMaxSnapshots(t *testing.T) {
	c := NewCollector(3)
	c.Reset()

	for i := 0; i < 5; i++ {
		c.saveSnapshot()
	}

	snapshots := c.GetSnapshots()

	if len(snapshots) > 3 {
		t.Errorf("快照数量不应超过最大值3, 实际为 %d", len(snapshots))
	}
}

func TestClearSnapshots(t *testing.T) {
	c := NewCollector(100)
	c.Reset()

	c.saveSnapshot()
	c.saveSnapshot()
	c.saveSnapshot()

	if c.GetSnapshotsCount() != 3 {
		t.Errorf("应有3个快照, 实际有 %d 个", c.GetSnapshotsCount())
	}

	c.ClearSnapshots()

	if c.GetSnapshotsCount() != 0 {
		t.Errorf("清空后快照数量应为0, 实际为 %d", c.GetSnapshotsCount())
	}
}

func TestConcurrentIncrement(t *testing.T) {
	c := NewCollector(100)
	c.Reset()

	var wg sync.WaitGroup
	numGoroutines := 100
	incrementsPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				c.IncrementConnections()
				c.IncrementRequests()
				c.AddBytesReceived(1)
				c.AddBytesSent(1)
			}
		}()
	}

	wg.Wait()

	snapshot := c.GetSnapshot()
	expected := int64(numGoroutines * incrementsPerGoroutine)

	if snapshot.TotalConnections != expected {
		t.Errorf("TotalConnections 应为 %d, 实际为 %d", expected, snapshot.TotalConnections)
	}

	if snapshot.Requests != expected {
		t.Errorf("Requests 应为 %d, 实际为 %d", expected, snapshot.Requests)
	}

	if snapshot.BytesReceived != expected {
		t.Errorf("BytesReceived 应为 %d, 实际为 %d", expected, snapshot.BytesReceived)
	}

	if snapshot.BytesSent != expected {
		t.Errorf("BytesSent 应为 %d, 实际为 %d", expected, snapshot.BytesSent)
	}
}

func TestConcurrentEnableDisable(t *testing.T) {
	c := NewCollector(100)
	c.Reset()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(enable bool) {
			defer wg.Done()
			if enable {
				c.Enable()
			} else {
				c.Disable()
			}
			_ = c.IsEnabled()
		}(i%2 == 0)
	}

	wg.Wait()
}

func TestDefaultCollector(t *testing.T) {
	c1 := DefaultCollector()
	c2 := DefaultCollector()

	if c1 != c2 {
		t.Error("DefaultCollector 应返回单例")
	}
}

func TestPackageFunctions(t *testing.T) {
	Enable()
	if !IsEnabled() {
		t.Error("全局 Enable() 后 IsEnabled() 应返回 true")
	}

	IncrementConnections()
	DecrementConnections()
	AddBytesReceived(100)
	AddBytesSent(50)
	IncrementRequests()
	IncrementErrors()
	RecordLatency(10 * time.Millisecond)

	stats := GetStats()
	if stats.TotalConnections != 1 {
		t.Errorf("TotalConnections 应为 1, 实际为 %d", stats.TotalConnections)
	}

	snapshot := GetSnapshot()
	if snapshot.TotalConnections != 1 {
		t.Errorf("快照 TotalConnections 应为 1, 实际为 %d", snapshot.TotalConnections)
	}

	_ = GetSnapshots()
}

func TestNegativeActiveConnections(t *testing.T) {
	c := NewCollector(100)
	c.Reset()

	for i := 0; i < 5; i++ {
		c.DecrementConnections()
	}

	snapshot := c.GetSnapshot()
	if snapshot.ActiveConnections != -5 {
		t.Errorf("ActiveConnections 应为 -5, 实际为 %d", snapshot.ActiveConnections)
	}
}

func TestSnapshotSchedulerStartStop(t *testing.T) {
	c := NewCollector(100)
	c.Reset()

	c.StartSnapshotScheduler(100 * time.Millisecond)
	c.StartSnapshotScheduler(100 * time.Millisecond)

	c.StopSnapshotScheduler()
	c.StopSnapshotScheduler()

	c.StartSnapshotScheduler(50 * time.Millisecond)
	time.Sleep(120 * time.Millisecond)
	c.StopSnapshotScheduler()

	if c.GetSnapshotsCount() < 1 {
		t.Errorf("快照调度器应生成至少1个快照, 实际为 %d", c.GetSnapshotsCount())
	}
}
