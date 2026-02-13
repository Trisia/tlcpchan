package proxy

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestParseProtocolType(t *testing.T) {
	tests := []struct {
		input    string
		expected ProtocolType
	}{
		{"tlcp", ProtocolTLCP},
		{"tls", ProtocolTLS},
		{"auto", ProtocolAuto},
		{"", ProtocolAuto},
		{"unknown", ProtocolAuto},
		{"TLCP", ProtocolAuto},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseProtocolType(tt.input)
			if result != tt.expected {
				t.Errorf("ParseProtocolType(%q) = %v, 期望 %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDetectProtocol(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected ProtocolType
	}{
		{
			name:     "TLCP ClientHello",
			data:     []byte{22, 3, 1, 1, 1, 0},
			expected: ProtocolTLCP,
		},
		{
			name:     "TLS 1.0 ClientHello",
			data:     []byte{22, 3, 1, 3, 1, 0},
			expected: ProtocolTLS,
		},
		{
			name:     "TLS 1.1 ClientHello",
			data:     []byte{22, 3, 2, 3, 1, 0},
			expected: ProtocolTLS,
		},
		{
			name:     "TLS 1.2 ClientHello",
			data:     []byte{22, 3, 3, 3, 1, 0},
			expected: ProtocolTLS,
		},
		{
			name:     "TLS 1.3 ClientHello",
			data:     []byte{22, 3, 4, 3, 1, 0},
			expected: ProtocolTLS,
		},
		{
			name:     "数据过短",
			data:     []byte{22, 3, 1, 1},
			expected: ProtocolTLS,
		},
		{
			name:     "非握手记录",
			data:     []byte{21, 3, 1, 1, 1, 0},
			expected: ProtocolTLS,
		},
		{
			name:     "空数据",
			data:     []byte{},
			expected: ProtocolTLS,
		},
		{
			name:     "TLCP版本标识0x0101",
			data:     []byte{22, 0, 0, 0x01, 0x01, 0},
			expected: ProtocolTLCP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectProtocol(tt.data)
			if result != tt.expected {
				t.Errorf("detectProtocol() = %v, 期望 %v", result, tt.expected)
			}
		})
	}
}

func TestDetectProtocolEdgeCases(t *testing.T) {
	t.Run("刚好5字节TLS", func(t *testing.T) {
		data := []byte{22, 3, 1, 0x03, 0x01}
		result := detectProtocol(data)
		if result != ProtocolTLS {
			t.Errorf("5字节TLS数据应返回 ProtocolTLS, got %v", result)
		}
	})

	t.Run("TLCP版本大小端", func(t *testing.T) {
		data := []byte{22, 0, 0, 0x01, 0x01, 0}
		result := detectProtocol(data)
		if result != ProtocolTLCP {
			t.Errorf("TLCP版本 0x0101 应返回 ProtocolTLCP")
		}
	})

	t.Run("TLS版本0x0301", func(t *testing.T) {
		data := []byte{22, 0, 0, 0x03, 0x01, 0}
		result := detectProtocol(data)
		if result != ProtocolTLS {
			t.Errorf("TLS版本 0x0301 应返回 ProtocolTLS")
		}
	})
}

func TestProtocolTypeConstants(t *testing.T) {
	if ProtocolAuto != 0 {
		t.Errorf("ProtocolAuto 应为 0, 实际为 %d", ProtocolAuto)
	}

	if ProtocolTLCP != 1 {
		t.Errorf("ProtocolTLCP 应为 1, 实际为 %d", ProtocolTLCP)
	}

	if ProtocolTLS != 2 {
		t.Errorf("ProtocolTLS 应为 2, 实际为 %d", ProtocolTLS)
	}
}

func TestAutoProtocolConnPeekData(t *testing.T) {
	conn := &autoProtocolConn{
		peeked: []byte{1, 2, 3, 4, 5},
	}

	buf := make([]byte, 10)
	n, err := conn.Read(buf)
	if err != nil {
		t.Errorf("Read() 不应返回错误: %v", err)
	}

	if n != 5 {
		t.Errorf("Read() 应返回 5, 实际返回 %d", n)
	}

	if !bytes.Equal(buf[:n], []byte{1, 2, 3, 4, 5}) {
		t.Errorf("读取的数据不匹配")
	}

	if len(conn.peeked) != 0 {
		t.Errorf("读取后 peeked 应为空")
	}
}

func TestAutoProtocolConnPartialPeekRead(t *testing.T) {
	conn := &autoProtocolConn{
		peeked: []byte{1, 2, 3, 4, 5},
	}

	buf := make([]byte, 3)
	n, err := conn.Read(buf)
	if err != nil {
		t.Errorf("Read() 不应返回错误: %v", err)
	}

	if n != 3 {
		t.Errorf("Read() 应返回 3, 实际返回 %d", n)
	}

	if !bytes.Equal(buf, []byte{1, 2, 3}) {
		t.Errorf("读取的数据不匹配")
	}

	if len(conn.peeked) != 2 {
		t.Errorf("peeked 应剩余 2 字节, 实际为 %d", len(conn.peeked))
	}
}

func TestAutoProtocolConnMultiplePartialReads(t *testing.T) {
	conn := &autoProtocolConn{
		peeked: []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}

	buf := make([]byte, 3)

	n, _ := conn.Read(buf)
	if n != 3 || !bytes.Equal(buf, []byte{1, 2, 3}) {
		t.Errorf("第一次读取失败")
	}

	n, _ = conn.Read(buf)
	if n != 3 || !bytes.Equal(buf, []byte{4, 5, 6}) {
		t.Errorf("第二次读取失败")
	}

	n, _ = conn.Read(buf)
	if n != 2 || !bytes.Equal(buf[:n], []byte{7, 8}) {
		t.Errorf("第三次读取失败")
	}
}

func TestAutoProtocolConnHandshakedRead(t *testing.T) {
	mockConn := &mockNetConn{}
	conn := &autoProtocolConn{
		Conn:       mockConn,
		handshaked: true,
		conn:       mockConn,
	}

	mockConn.data = []byte{10, 20, 30}
	buf := make([]byte, 10)
	n, err := conn.Read(buf)
	if err != nil {
		t.Errorf("Read() 不应返回错误: %v", err)
	}

	if n != 3 {
		t.Errorf("Read() 应返回 3, 实际返回 %d", n)
	}
}

func TestAutoProtocolConnHandshakedWrite(t *testing.T) {
	mockConn := &mockNetConn{}
	conn := &autoProtocolConn{
		Conn:       mockConn,
		handshaked: true,
		conn:       mockConn,
	}

	data := []byte{1, 2, 3, 4, 5}
	n, err := conn.Write(data)
	if err != nil {
		t.Errorf("Write() 不应返回错误: %v", err)
	}

	if n != 5 {
		t.Errorf("Write() 应返回 5, 实际返回 %d", n)
	}
}

func TestAutoProtocolConnNotHandshakedWrite(t *testing.T) {
	mockConn := &mockNetConn{}
	conn := &autoProtocolConn{
		Conn:       mockConn,
		handshaked: false,
	}

	data := []byte{1, 2, 3, 4, 5}
	n, err := conn.Write(data)
	if err != nil {
		t.Errorf("Write() 不应返回错误: %v", err)
	}

	if n != 5 {
		t.Errorf("Write() 应返回 5, 实际返回 %d", n)
	}
}

func TestAutoProtocolConnCloseWithHandshaked(t *testing.T) {
	mockConn := &mockNetConn{}
	conn := &autoProtocolConn{
		Conn:       mockConn,
		handshaked: true,
		conn:       mockConn,
	}

	err := conn.Close()
	if err != nil {
		t.Errorf("Close() 不应返回错误: %v", err)
	}

	if !mockConn.closed {
		t.Error("底层连接应被关闭")
	}
}

func TestAutoProtocolConnCloseWithoutHandshaked(t *testing.T) {
	mockConn := &mockNetConn{}
	conn := &autoProtocolConn{
		Conn:       mockConn,
		handshaked: false,
	}

	err := conn.Close()
	if err != nil {
		t.Errorf("Close() 不应返回错误: %v", err)
	}

	if !mockConn.closed {
		t.Error("底层连接应被关闭")
	}
}

func TestAutoProtocolConnHandshakeTwice(t *testing.T) {
	conn := &autoProtocolConn{
		handshaked: true,
	}

	err := conn.Handshake()
	if err != nil {
		t.Errorf("重复握手不应返回错误: %v", err)
	}
}

func TestNewAutoProtocolListener(t *testing.T) {
	mockListener := &mockListener{}
	l := NewAutoProtocolListener(mockListener, nil, nil)

	if l == nil {
		t.Fatal("NewAutoProtocolListener() 返回 nil")
	}

	if l.Listener == nil {
		t.Error("Listener 未正确设置")
	}
}

func TestNewAutoProtocolConn(t *testing.T) {
	mockConn := &mockNetConn{}
	conn := newAutoProtocolConn(mockConn, nil, nil)

	if conn == nil {
		t.Fatal("newAutoProtocolConn() 返回 nil")
	}

	if conn.Conn == nil {
		t.Error("Conn 未正确设置")
	}
}

type mockNetConn struct {
	data   []byte
	closed bool
}

func (m *mockNetConn) Read(b []byte) (n int, err error) {
	n = copy(b, m.data)
	m.data = m.data[n:]
	return n, nil
}

func (m *mockNetConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (m *mockNetConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockNetConn) LocalAddr() net.Addr {
	return nil
}

func (m *mockNetConn) RemoteAddr() net.Addr {
	return nil
}

func (m *mockNetConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockNetConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockNetConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type mockListener struct{}

func (m *mockListener) Accept() (net.Conn, error) {
	return &mockNetConn{}, nil
}

func (m *mockListener) Close() error {
	return nil
}

func (m *mockListener) Addr() net.Addr {
	return nil
}
