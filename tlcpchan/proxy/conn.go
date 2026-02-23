package proxy

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/stats"
)

type ConnHandler struct {
	stats  *stats.Collector
	logger *logger.Logger
}

func NewConnHandler(stats *stats.Collector) *ConnHandler {
	return &ConnHandler{
		stats:  stats,
		logger: logger.Default(),
	}
}

func (h *ConnHandler) Pipe(ctx context.Context, clientConn, targetConn net.Conn) (received int64, sent int64, err error) {
	var wg sync.WaitGroup
	wg.Add(2)

	var clientToTargetErr error
	var targetToClientErr error

	go func() {
		defer wg.Done()
		var n int64
		n, clientToTargetErr = io.Copy(targetConn, clientConn)
		sent = n
		if h.stats != nil {
			h.stats.AddBytesSent(n)
		}
	}()

	go func() {
		defer wg.Done()
		var n int64
		n, targetToClientErr = io.Copy(clientConn, targetConn)
		received = n
		if h.stats != nil {
			h.stats.AddBytesReceived(n)
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		clientConn.Close()
		targetConn.Close()
		<-done
	case <-done:
	}

	if clientToTargetErr != nil && !isNormalClose(clientToTargetErr) {
		return received, sent, clientToTargetErr
	}
	if targetToClientErr != nil && !isNormalClose(targetToClientErr) {
		return received, sent, targetToClientErr
	}

	return received, sent, nil
}

func isNormalClose(err error) bool {
	if err == nil {
		return false
	}
	return err == io.EOF || err == io.ErrUnexpectedEOF ||
		err.Error() == "use of closed network connection"
}
