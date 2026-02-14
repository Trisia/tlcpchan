package instance

// Status 实例运行状态
type Status string

const (
	// StatusCreated 已创建，未启动
	StatusCreated Status = "created"
	// StatusRunning 运行中
	StatusRunning Status = "running"
	// StatusStopped 已停止
	StatusStopped Status = "stopped"
	// StatusError 错误状态
	StatusError Status = "error"
)

// InstanceType 实例类型
type InstanceType string

const (
	// TypeServer TCP服务端代理，接收TLCP/TLS连接并转发到目标服务
	TypeServer InstanceType = "server"
	// TypeClient TCP客户端代理，接收普通TCP连接并以TLCP/TLS连接目标服务
	TypeClient InstanceType = "client"
	// TypeHTTPServer HTTP服务端代理，处理HTTP/HTTPS请求
	TypeHTTPServer InstanceType = "http-server"
	// TypeHTTPClient HTTP客户端代理，发起HTTP/HTTPS请求
	TypeHTTPClient InstanceType = "http-client"
)

// ParseInstanceType 解析实例类型字符串
// 参数:
//   - s: 类型字符串，如 "server", "client", "http-server", "http-client"
//
// 返回:
//   - InstanceType: 实例类型，无法识别时返回 TypeServer
func ParseInstanceType(s string) InstanceType {
	switch s {
	case "server":
		return TypeServer
	case "client":
		return TypeClient
	case "http-server":
		return TypeHTTPServer
	case "http-client":
		return TypeHTTPClient
	default:
		return TypeServer
	}
}
