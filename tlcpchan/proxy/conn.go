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

type ConnHandler struct {
	stats       *stats.Collector
	logger      *logger.Logger
	readTimeout time.Duration
}

func NewConnHandler() *ConnHandler {
	return &ConnHandler{
		stats:  stats.DefaultCollector(),
		logger: logger.Default(),
	}
}

func (h *ConnHandler) SetReadTimeout(d time.Duration) {
	h.readTimeout = d
}

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
		n, err := io.Copy(target, client)
		if n > 0 {
			h.stats.AddBytesSent(n)
			sent.Add(n)
		}
		recordError(err)
		target.Close()
	}()

	go func() {
		defer wg.Done()
		n, err := io.Copy(client, target)
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

type ConnWrapper struct {
	net.Conn
	readTimeout  time.Duration
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
