package proxy

import (
	"net"
	"strings"
)

type Variables struct {
	RemoteAddr   string
	RemoteIP     string
	RemotePort   string
	ServerAddr   string
	ServerIP     string
	ServerPort   string
	TargetAddr   string
	TargetIP     string
	TargetPort   string
	Protocol     string
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
