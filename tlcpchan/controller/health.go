package controller

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/Trisia/tlcpchan/cert"
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/instance"
)

// HealthController 健康检测控制器
type HealthController struct {
	instanceMgr *instance.Manager
	certMgr     *cert.Manager
}

// NewHealthController 创建健康检测控制器
// 参数:
//   - instanceMgr: 实例管理器
//   - certMgr: 证书管理器
//
// 返回:
//   - *HealthController: 健康检测控制器实例
func NewHealthController(instanceMgr *instance.Manager, certMgr *cert.Manager) *HealthController {
	return &HealthController{
		instanceMgr: instanceMgr,
		certMgr:     certMgr,
	}
}

// HealthCheckRequest 健康检测请求
type HealthCheckRequest struct {
	// FullHandshake 是否执行完整握手测试
	FullHandshake bool `json:"full_handshake"`
}

// HealthCheckResponse 健康检测结果
type HealthCheckResponse struct {
	// Success 是否成功
	Success bool `json:"success"`
	// LatencyMs 连接延迟，单位: 毫秒
	LatencyMs float64 `json:"latency_ms"`
	// Error 错误信息
	Error string `json:"error,omitempty"`
	// TLCPInfo TLCP检测结果
	TLCPInfo *ProtocolHealthInfo `json:"tlcp_info,omitempty"`
	// TLSInfo TLS检测结果
	TLSInfo *ProtocolHealthInfo `json:"tls_info,omitempty"`
}

// ProtocolHealthInfo 协议健康信息
type ProtocolHealthInfo struct {
	// Success 是否成功
	Success bool `json:"success"`
	// LatencyMs 握手延迟，单位: 毫秒
	LatencyMs float64 `json:"latency_ms"`
	// CertValid 证书是否有效
	CertValid bool `json:"cert_valid"`
	// CertExpiry 证书过期时间
	CertExpiry string `json:"cert_expiry,omitempty"`
	// CertDaysRemaining 证书剩余天数
	CertDaysRemaining int `json:"cert_days_remaining,omitempty"`
	// Error 错误信息
	Error string `json:"error,omitempty"`
}

/**
 * @api {post} /api/v1/instances/:name/health 健康检测（POST）
 * @apiName CheckHealthPost
 * @apiGroup Health
 * @apiVersion 1.0.0
 *
 * @apiDescription 执行实例的健康检测，支持完整握手测试
 *
 * @apiParam {String} name 实例名称（路径参数）
 * @apiParam {Boolean} [full_handshake=false] 是否执行完整握手测试
 *
 * @apiSuccess {Boolean} success 是否成功
 * @apiSuccess {Number} latency_ms 连接延迟（毫秒）
 * @apiSuccess {String} [error] 错误信息
 * @apiSuccess {Object} [tlcp_info] TLCP检测结果
 * @apiSuccess {Object} [tls_info] TLS检测结果
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "success": true,
 *       "latency_ms": 5.2,
 *       "tlcp_info": {
 *         "success": true,
 *         "latency_ms": 5.2,
 *         "cert_valid": true,
 *         "cert_expiry": "2025-12-31T23:59:59Z",
 *         "cert_days_remaining": 365
 *       }
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     实例名称不能为空
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     实例不存在
 */

/**
 * @api {get} /api/v1/instances/:name/health 健康检测（GET）
 * @apiName CheckHealthGet
 * @apiGroup Health
 * @apiVersion 1.0.0
 *
 * @apiDescription 执行实例的健康检测（简化版，不执行完整握手）
 *
 * @apiParam {String} name 实例名称（路径参数）
 *
 * @apiSuccess {Boolean} success 是否成功
 * @apiSuccess {Number} latency_ms 连接延迟（毫秒）
 * @apiSuccess {String} [error] 错误信息
 * @apiSuccess {Object} [tlcp_info] TLCP检测结果
 * @apiSuccess {Object} [tls_info] TLS检测结果
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "success": true,
 *       "latency_ms": 5.2
 *     }
 */
// Check 执行实例健康检测
// POST /api/v1/instances/:name/health
func (c *HealthController) Check(w http.ResponseWriter, r *http.Request) {
	name := extractPathParam(r.URL.Path, "/api/v1/instances/", "/health")
	if name == "" {
		WriteError(w, http.StatusBadRequest, "实例名称不能为空")
		return
	}

	inst, exists := c.instanceMgr.Get(name)
	if !exists {
		WriteError(w, http.StatusNotFound, "实例不存在")
		return
	}

	cfg := inst.Config()
	if cfg == nil {
		WriteError(w, http.StatusInternalServerError, "无法获取实例配置")
		return
	}

	// 解析请求
	var req HealthCheckRequest
	if r.Method == "POST" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			req.FullHandshake = false
		}
	}

	result := &HealthCheckResponse{}

	// 根据协议类型执行检测
	switch cfg.Protocol {
	case "tlcp":
		result.TLCPInfo = c.checkTLCP(cfg, req.FullHandshake)
		result.Success = result.TLCPInfo.Success
		result.LatencyMs = result.TLCPInfo.LatencyMs
		if result.TLCPInfo.Error != "" {
			result.Error = result.TLCPInfo.Error
		}
	case "tls":
		result.TLSInfo = c.checkTLS(cfg, req.FullHandshake)
		result.Success = result.TLSInfo.Success
		result.LatencyMs = result.TLSInfo.LatencyMs
		if result.TLSInfo.Error != "" {
			result.Error = result.TLSInfo.Error
		}
	default:
		// auto模式，两种协议都检测
		result.TLCPInfo = c.checkTLCP(cfg, req.FullHandshake)
		result.TLSInfo = c.checkTLS(cfg, req.FullHandshake)
		result.Success = result.TLCPInfo.Success || result.TLSInfo.Success
		if result.TLCPInfo.LatencyMs > 0 {
			result.LatencyMs = result.TLCPInfo.LatencyMs
		} else if result.TLSInfo.LatencyMs > 0 {
			result.LatencyMs = result.TLSInfo.LatencyMs
		}
		if result.TLCPInfo.Error != "" && result.TLSInfo.Error != "" {
			result.Error = "TLCP: " + result.TLCPInfo.Error + "; TLS: " + result.TLSInfo.Error
		}
	}

	WriteJSON(w, http.StatusOK, result)
}

// checkTLCP 检测TLCP连接健康状态
func (c *HealthController) checkTLCP(cfg *config.InstanceConfig, fullHandshake bool) *ProtocolHealthInfo {
	info := &ProtocolHealthInfo{}

	start := time.Now()

	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	conn, err := dialer.Dial("tcp", cfg.Target)
	if err != nil {
		info.Error = "连接失败: " + err.Error()
		return info
	}
	defer conn.Close()

	if !fullHandshake {
		info.Success = true
		info.LatencyMs = float64(time.Since(start).Milliseconds())
		return info
	}

	// 执行完整TLCP握手
	tlcpCfg := &tlcp.Config{
		InsecureSkipVerify: cfg.TLCP.InsecureSkipVerify,
	}

	tlcpConn := tlcp.Client(conn, tlcpCfg)
	if err := tlcpConn.Handshake(); err != nil {
		info.Error = "TLCP握手失败: " + err.Error()
		info.LatencyMs = float64(time.Since(start).Milliseconds())
		return info
	}

	info.Success = true
	info.LatencyMs = float64(time.Since(start).Milliseconds())

	// 获取证书信息
	state := tlcpConn.ConnectionState()
	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]
		info.CertValid = time.Now().Before(cert.NotAfter)
		info.CertExpiry = cert.NotAfter.Format(time.RFC3339)
		info.CertDaysRemaining = int(time.Until(cert.NotAfter).Hours() / 24)
	}

	return info
}

// checkTLS 检测TLS连接健康状态
func (c *HealthController) checkTLS(cfg *config.InstanceConfig, fullHandshake bool) *ProtocolHealthInfo {
	info := &ProtocolHealthInfo{}

	start := time.Now()

	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	conn, err := dialer.Dial("tcp", cfg.Target)
	if err != nil {
		info.Error = "连接失败: " + err.Error()
		return info
	}
	defer conn.Close()

	if !fullHandshake {
		info.Success = true
		info.LatencyMs = float64(time.Since(start).Milliseconds())
		return info
	}

	// 执行完整TLS握手
	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
	}

	if cfg.SNI != "" {
		tlsCfg.ServerName = cfg.SNI
	}

	tlsConn := tls.Client(conn, tlsCfg)
	if err := tlsConn.Handshake(); err != nil {
		info.Error = "TLS握手失败: " + err.Error()
		info.LatencyMs = float64(time.Since(start).Milliseconds())
		return info
	}

	info.Success = true
	info.LatencyMs = float64(time.Since(start).Milliseconds())

	// 获取证书信息
	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]
		info.CertValid = time.Now().Before(cert.NotAfter)
		info.CertExpiry = cert.NotAfter.Format(time.RFC3339)
		info.CertDaysRemaining = int(time.Until(cert.NotAfter).Hours() / 24)
	}

	return info
}

// extractPathParam 从URL路径中提取参数
func extractPathParam(path, prefix, suffix string) string {
	if len(path) <= len(prefix)+len(suffix) {
		return ""
	}
	return path[len(prefix) : len(path)-len(suffix)]
}

// RegisterRoutes 注册健康检测路由
// 参数:
//   - router: HTTP路由器
func (c *HealthController) RegisterRoutes(router *Router) {
	router.POST("/api/v1/instances/:name/health", c.Check)
	router.GET("/api/v1/instances/:name/health", c.Check)
}
