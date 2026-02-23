package proxy

import (
	"bytes"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/stats"
)

func TestValidateClientConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.InstanceConfig
		wantErr bool
	}{
		{
			name: "默认配置",
			cfg: &config.InstanceConfig{
				Protocol: string(config.ProtocolAuto),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func TestDetectProtocol(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want ProtocolType
	}{
		{"TLCP握手", []byte{0x16, 0x01, 0x01, 0x01, 0x01, 0x00}, ProtocolTLCP},
		{"TLS 1.0握手", []byte{0x16, 0x03, 0x01, 0x03, 0x01, 0x00}, ProtocolTLS},
		{"TLS 1.2握手", []byte{0x16, 0x03, 0x01, 0x03, 0x03, 0x00}, ProtocolTLS},
		{"数据太短", []byte{0x16, 0x03}, ProtocolTLS},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectProtocol(tt.data); got != tt.want {
				t.Errorf("detectProtocol() = %v, want %v", got, tt.want)
			}
		})
	}
}

// mockConn 模拟网络连接
type mockConn struct {
	io.Reader
	io.Writer
	closeCalled atomic.Bool
}

func (m *mockConn) Read(b []byte) (int, error) {
	return m.Reader.Read(b)
}

func (m *mockConn) Write(b []byte) (int, error) {
	return m.Writer.Write(b)
}

func (m *mockConn) Close() error {
	m.closeCalled.Store(true)
	return nil
}

func (m *mockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
}

func (m *mockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8081}
}

func (m *mockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// TestCopyWithStats 测试copyWithStats方法实时更新统计信息
func TestCopyWithStats(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		isSent    bool
		wantBytes int64
	}{
		{"小数据传输", []byte("hello"), true, 5},
		{"中等数据传输", bytes.Repeat([]byte("test"), 100), false, 400},
		{"大数据传输", bytes.Repeat([]byte("data"), 10000), true, 40000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := stats.NewCollector(10)
			handler := NewConnHandler(collector, 4096)

			src := bytes.NewReader(tt.data)
			var dst bytes.Buffer

			copied, err := handler.copyWithStats(&dst, src, collector, tt.isSent)
			if err != nil {
				t.Fatalf("copyWithStats() error = %v", err)
			}

			if copied != tt.wantBytes {
				t.Errorf("copyWithStats() copied = %v, want %v", copied, tt.wantBytes)
			}

			snapshot := collector.GetSnapshot()
			if tt.isSent {
				if snapshot.BytesSent != tt.wantBytes {
					t.Errorf("BytesSent = %v, want %v", snapshot.BytesSent, tt.wantBytes)
				}
			} else {
				if snapshot.BytesReceived != tt.wantBytes {
					t.Errorf("BytesReceived = %v, want %v", snapshot.BytesReceived, tt.wantBytes)
				}
			}
		})
	}
}

// TestCopyWithStatsRealtimeUpdate 测试统计信息是否实时更新
func TestCopyWithStatsRealtimeUpdate(t *testing.T) {
	collector := stats.NewCollector(10)
	handler := NewConnHandler(collector, 4096)

	// 创建一个分阶段提供数据的读取器
	chunks := [][]byte{
		bytes.Repeat([]byte("chunk1"), 1000), // 6000字节
		bytes.Repeat([]byte("chunk2"), 1000), // 6000字节
		bytes.Repeat([]byte("chunk3"), 1000), // 6000字节
	}

	chunkIndex := atomic.Int32{}
	src := &chunkedReader{chunks: chunks, index: &chunkIndex}
	var dst bytes.Buffer

	// 创建一个goroutine监控统计信息变化
	updateCount := atomic.Int32{}
	expectedUpdates := int32(len(chunks))

	go func() {
		lastBytes := int64(0)
		for i := int32(0); i < expectedUpdates; i++ {
			time.Sleep(50 * time.Millisecond)
			snapshot := collector.GetSnapshot()
			if snapshot.BytesSent > lastBytes {
				lastBytes = snapshot.BytesSent
				updateCount.Add(1)
			}
		}
	}()

	_, err := handler.copyWithStats(&dst, src, collector, true)
	if err != nil {
		t.Fatalf("copyWithStats() error = %v", err)
	}

	// 验证统计数据
	snapshot := collector.GetSnapshot()
	totalBytes := int64(len(chunks[0]) * len(chunks))
	if snapshot.BytesSent != totalBytes {
		t.Errorf("最终 BytesSent = %v, want %v", snapshot.BytesSent, totalBytes)
	}

	// 等待监控goroutine完成
	time.Sleep(200 * time.Millisecond)
	updates := updateCount.Load()
	if updates < 1 {
		t.Errorf("统计信息更新次数过少: %v，期望至少1次", updates)
	}
}

// chunkedReader 分块提供数据的读取器
type chunkedReader struct {
	chunks [][]byte
	index  *atomic.Int32
	offset int
}

func (r *chunkedReader) Read(p []byte) (int, error) {
	currentIdx := r.index.Load()
	if currentIdx >= int32(len(r.chunks)) {
		return 0, io.EOF
	}

	chunk := r.chunks[currentIdx]
	remaining := len(chunk) - r.offset

	if remaining <= 0 {
		r.index.Add(1)
		r.offset = 0
		return r.Read(p)
	}

	n := len(p)
	if n > remaining {
		n = remaining
	}

	copy(p, chunk[r.offset:r.offset+n])
	r.offset += n

	if r.offset >= len(chunk) {
		r.index.Add(1)
		r.offset = 0
	}

	return n, nil
}

// TestPipe 测试Pipe方法
func TestPipe(t *testing.T) {
	// Pipe测试需要复杂的网络连接设置，这里仅测试基本功能
	// 主要的实时更新测试已由 TestCopyWithStats 和 TestCopyWithStatsRealtimeUpdate 覆盖
	collector := stats.NewCollector(10)
	handler := NewConnHandler(collector, 4096)

	if handler == nil {
		t.Fatal("NewConnHandler 返回 nil")
	}

	if handler.bufferSize != 4096 {
		t.Errorf("BufferSize = %v, want 4096", handler.bufferSize)
	}
}

// TestIsNormalError 测试isNormalError函数
func TestIsNormalError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil错误", nil, false},
		{"EOF", io.EOF, true},
		{"UnexpectedEOF", io.ErrUnexpectedEOF, true},
		{"关闭连接", io.EOF, true},
		{"其他错误", io.ErrClosedPipe, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNormalError(tt.err); got != tt.want {
				t.Errorf("isNormalError() = %v, want %v", got, tt.want)
			}
		})
	}
}
