package proxy

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/stats"
)

type ConnHandler struct {
	stats      *stats.Collector
	logger     *logger.Logger
	bufferSize int
}

func NewConnHandler(stats *stats.Collector, bufferSize int) *ConnHandler {
	if bufferSize <= 0 {
		bufferSize = 4096
	}
	return &ConnHandler{
		stats:      stats,
		logger:     logger.Default(),
		bufferSize: bufferSize,
	}
}

// copyWithStats 从src复制数据到dst，并在每次写入时更新统计信息
// 参数:
//   - dst: 目标写入器
//   - src: 源读取器
//   - stats: 统计收集器，可为nil
//   - isSent: true表示发送统计，false表示接收统计
//
// 返回:
//   - int64: 复制的字节数
//   - error: 错误信息
func (h *ConnHandler) copyWithStats(dst io.Writer, src io.Reader, stats *stats.Collector, isSent bool) (int64, error) {
	buf := make([]byte, h.bufferSize)
	var written int64

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
				// 实时更新统计信息
				if stats != nil {
					if isSent {
						stats.AddBytesSent(int64(nw))
					} else {
						stats.AddBytesReceived(int64(nw))
					}
				}
			}
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if er != nil {
			if er.Error() == "EOF" {
				return written, nil
			}
			if errors.Is(er, io.EOF) {
				return written, nil
			}
			return written, er
		}
	}
}

// Pipe 在clientConn和targetConn之间建立双向数据管道
// 参数:
//   - ctx: 上下文，用于取消操作
//   - clientConn: 客户端连接
//   - targetConn: 目标服务器连接
//
// 返回:
//   - int64: 接收的总字节数
//   - int64: 发送的总字节数
//   - error: 错误信息
//
// 注意: 使用自定义复制函数实现统计信息实时更新
func (h *ConnHandler) Pipe(ctx context.Context, clientConn, targetConn net.Conn) (received int64, sent int64, err error) {
	var wg sync.WaitGroup
	wg.Add(2)

	var clientToTargetErr error
	var targetToClientErr error

	go func() {
		defer wg.Done()
		var n int64
		n, clientToTargetErr = h.copyWithStats(targetConn, clientConn, h.stats, true)
		sent = n
	}()

	go func() {
		defer wg.Done()
		var n int64
		n, targetToClientErr = h.copyWithStats(clientConn, targetConn, h.stats, false)
		received = n
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

	if clientToTargetErr != nil && !isNormalError(clientToTargetErr) {
		return received, sent, clientToTargetErr
	}
	if targetToClientErr != nil && !isNormalError(targetToClientErr) {
		return received, sent, targetToClientErr
	}

	return received, sent, nil
}

// isNormalError 判断是否为正常的连接关闭错误
// 参数:
//   - err: 错误信息
//
// 返回:
//   - bool: true表示是正常的关闭错误，false表示是异常错误
func isNormalError(err error) bool {
	if err == nil {
		return false
	}
	return err == io.EOF || err == io.ErrUnexpectedEOF ||
		err.Error() == "use of closed network connection"
}
