package proxy

import (
	"net"
	"strings"
)

// Variables 变量集合，用于配置文件中的变量替换
// 支持在HTTP头配置中使用 $remote_addr, $remote_ip 等变量
type Variables struct {
	// RemoteAddr 客户端地址，格式: "host:port"
	RemoteAddr string
	// RemoteIP 客户端IP地址
	RemoteIP string
	// RemotePort 客户端端口
	RemotePort string
	// ServerAddr 服务端地址，格式: "host:port"
	ServerAddr string
	// ServerIP 服务端IP地址
	ServerIP string
	// ServerPort 服务端端口
	ServerPort string
	// TargetAddr 目标地址，格式: "host:port"
	TargetAddr string
	// TargetIP 目标IP地址
	TargetIP string
	// TargetPort 目标端口
	TargetPort string
	// Protocol 协议类型
	Protocol string
	// InstanceName 实例名称
	InstanceName string
}

func ExtractVariables(remoteAddr, serverAddr, targetAddr, protocol, instanceName string) *Variables {
	v := &Variables{
		RemoteAddr:   remoteAddr,
		ServerAddr:   serverAddr,
		TargetAddr:   targetAddr,
		Protocol:     protocol,
		InstanceName: instanceName,
	}

	v.RemoteIP, v.RemotePort = splitHostPort(remoteAddr)
	v.ServerIP, v.ServerPort = splitHostPort(serverAddr)
	v.TargetIP, v.TargetPort = splitHostPort(targetAddr)

	return v
}

func ExtractVariablesFromConn(remoteConn, targetConn net.Conn, protocol, instanceName string) *Variables {
	var remoteAddr, serverAddr, targetAddr string

	if remoteConn != nil {
		remoteAddr = remoteConn.RemoteAddr().String()
		serverAddr = remoteConn.LocalAddr().String()
	}

	if targetConn != nil {
		targetAddr = targetConn.RemoteAddr().String()
	}

	return ExtractVariables(remoteAddr, serverAddr, targetAddr, protocol, instanceName)
}

func splitHostPort(addr string) (host, port string) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		if strings.Contains(addr, ":") {
			lastColon := strings.LastIndex(addr, ":")
			if lastColon > 0 {
				return addr[:lastColon], addr[lastColon+1:]
			}
		}
		return addr, ""
	}
	return host, port
}

func (v *Variables) Replace(s string) string {
	s = strings.ReplaceAll(s, "$remote_addr", v.RemoteAddr)
	s = strings.ReplaceAll(s, "$remote_ip", v.RemoteIP)
	s = strings.ReplaceAll(s, "$remote_port", v.RemotePort)
	s = strings.ReplaceAll(s, "$server_addr", v.ServerAddr)
	s = strings.ReplaceAll(s, "$server_ip", v.ServerIP)
	s = strings.ReplaceAll(s, "$server_port", v.ServerPort)
	s = strings.ReplaceAll(s, "$target_addr", v.TargetAddr)
	s = strings.ReplaceAll(s, "$target_ip", v.TargetIP)
	s = strings.ReplaceAll(s, "$target_port", v.TargetPort)
	s = strings.ReplaceAll(s, "$protocol", v.Protocol)
	s = strings.ReplaceAll(s, "$instance", v.InstanceName)
	return s
}

func (v *Variables) ReplaceMap(m map[string]string) map[string]string {
	result := make(map[string]string, len(m))
	for k, val := range m {
		result[k] = v.Replace(val)
	}
	return result
}
