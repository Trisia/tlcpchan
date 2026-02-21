package connection

import (
	"encoding/json"
	"sync"

	"github.com/Trisia/tlcpchan/mcp/protocol"
	"github.com/gorilla/websocket"
)

// Connection MCP连接
type Connection struct {
	ws     *websocket.Conn
	id     string
	mu     sync.Mutex
	closed bool
}

// NewConnection 创建新连接
func NewConnection(ws *websocket.Conn, id string) *Connection {
	return &Connection{
		ws: ws,
		id: id,
	}
}

// ID 获取连接ID
func (c *Connection) ID() string {
	return c.id
}

// ReadRequest 读取请求
func (c *Connection) ReadRequest() (*protocol.Request, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, websocket.ErrCloseSent
	}

	_, data, err := c.ws.ReadMessage()
	if err != nil {
		return nil, err
	}

	var req protocol.Request
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

// WriteResponse 写入响应
func (c *Connection) WriteResponse(resp *protocol.Response) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return websocket.ErrCloseSent
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	return c.ws.WriteMessage(websocket.TextMessage, data)
}

// Close 关闭连接
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	return c.ws.Close()
}

// IsClosed 检查连接是否已关闭
func (c *Connection) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}
