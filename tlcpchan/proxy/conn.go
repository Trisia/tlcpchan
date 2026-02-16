package proxy

import (
	"context"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/stats"
)

// bufferPool 内存池，用于io.Copy的缓冲区复用
var bufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 4*1024)
		return &buf
	},
}

// getBuffer 从内存池获取缓冲区
// 返回:
//   - *[]byte: 缓冲区指针
func getBuffer() *[]byte {
	return bufferPool.Get().(*[]byte)
}

// putBuffer 将缓冲区归还到内存池
// 参数:
//   - buf: 缓冲区指针
func putBuffer(buf *[]byte) {
	bufferPool.Put(buf)
}

// copyWithPool 使用内存池的拷贝
// 参数:
//   - dst: 目标Writer
//   - src: 源Reader
//
// 返回:
//   - int64: 拷贝的字节数
//   - error: 拷贝过程中的错误
func copyWithPool(dst io.Writer, src io.Reader) (int64, error) {
	buf := getBuffer()
	defer putBuffer(buf)
	return io.CopyBuffer(dst, src, *buf)
}

// ConnHandler 连接处理器，负责双向数据转发
type ConnHandler struct {
	// stats 统计收集器
	stats  *stats.Collector
	logger *logger.Logger
	// readTimeout 读取超时时间，单位: 纳秒
	readTimeout time.Duration
}

// NewConnHandler 创建新的连接处理器
// 返回:
//   - *ConnHandler: 连接处理器实例
func NewConnHandler() *ConnHandler {
	return &ConnHandler{
		stats:  stats.DefaultCollector(),
		logger: logger.Default(),
	}
}

func (h *ConnHandler) SetReadTimeout(d time.Duration) {
	h.readTimeout = d
}

// Pipe 双向数据转发，将客户端和目标服务之间的数据进行双向拷贝
// 参数:
//   - ctx: 上下文，用于控制转发终止
//   - client: 客户端连接
//   - target: 目标服务连接
//
// 返回:
//   - int64: 从目标接收的字节数
//   - int64: 发送到目标的字节数
//   - error: 转发过程中的错误（不含EOF和连接关闭）
//
// 注意: 该方法会阻塞直到连接关闭或ctx取消
func (h *ConnHandler) Pipe(ctx context.Context, client, target net.Conn) (int64, int64, error) {
	var received, sent atomic.Int64
	var wg sync.WaitGroup
	var errs []error
	var errsMu sync.Mutex

	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		client.Close()
		target.Close()
	}()

	recordError := func(err error) {
		if err != nil && err != io.EOF && !isClosedErr(err) {
			errsMu.Lock()
			errs = append(errs, err)
			errsMu.Unlock()
		}
	}

	wg.Add(2)

	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				h.logger.Error("数据转发panic(客户端->目标): %v", r)
			}
		}()
		n, err := copyWithPool(target, client)
		if n > 0 {
			h.stats.AddBytesSent(n)
			sent.Add(n)
		}
		recordError(err)
		target.Close()
	}()

	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				h.logger.Error("数据转发panic(目标->客户端): %v", r)
			}
		}()
		n, err := copyWithPool(client, target)
		if n > 0 {
			h.stats.AddBytesReceived(n)
			received.Add(n)
		}
		recordError(err)
		client.Close()
	}()

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}

	if len(errs) > 0 {
		return received.Load(), sent.Load(), errs[0]
	}
	return received.Load(), sent.Load(), nil
}

func (h *ConnHandler) PipeWithTimeout(ctx context.Context, client, target net.Conn, timeout time.Duration) (int64, int64, error) {
	if timeout > 0 {
		if tc, ok := client.(*net.TCPConn); ok {
			tc.SetDeadline(time.Now().Add(timeout))
		}
		if tc, ok := target.(*net.TCPConn); ok {
			tc.SetDeadline(time.Now().Add(timeout))
		}
	}
	return h.Pipe(ctx, client, target)
}

func isClosedErr(err error) bool {
	if err == nil {
		return false
	}
	if err == io.EOF {
		return true
	}
	if opErr, ok := err.(*net.OpError); ok {
		return opErr.Err.Error() == "use of closed network connection"
	}
	return false
}

// ConnWrapper 连接包装器，提供读写超时控制
type ConnWrapper struct {
	net.Conn
	// readTimeout 读取超时时间，单位: 纳秒
	readTimeout time.Duration
	// writeTimeout 写入超时时间，单位: 纳秒
	writeTimeout time.Duration
}

func WrapConn(conn net.Conn) *ConnWrapper {
	return &ConnWrapper{Conn: conn}
}

func (c *ConnWrapper) SetTimeouts(read, write time.Duration) {
	c.readTimeout = read
	c.writeTimeout = write
}

func (c *ConnWrapper) Read(b []byte) (n int, err error) {
	if c.readTimeout > 0 {
		c.Conn.SetReadDeadline(time.Now().Add(c.readTimeout))
	}
	return c.Conn.Read(b)
}

func (c *ConnWrapper) Write(b []byte) (n int, err error) {
	if c.writeTimeout > 0 {
		c.Conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}
	return c.Conn.Write(b)
}

// GracefulClose 优雅关闭连接
// 参数:
//   - conn: 待关闭的连接
//   - timeout: 关闭超时时间，超时后强制关闭
//
// 注意: 对于TCP连接会先关闭读端，等待超时后完全关闭
func GracefulClose(conn net.Conn, timeout time.Duration) {
	if conn == nil {
		return
	}
	if timeout > 0 {
		conn.SetDeadline(time.Now().Add(timeout))
	}
	if tc, ok := conn.(*net.TCPConn); ok {
		tc.CloseRead()
	}
	conn.Close()
}
