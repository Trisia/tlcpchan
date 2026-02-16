package controller

import (
	"net/http"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
)

type ConfigController struct {
	configPath string
	cfg        *config.Config
	log        *logger.Logger
}

func NewConfigController(cfg *config.Config, configPath string) *ConfigController {
	return &ConfigController{
		configPath: configPath,
		cfg:        cfg,
		log:        logger.Default(),
	}
}

/**
 * @api {get} /api/v1/config 获取当前配置
 * @apiName GetConfig
 * @apiGroup Config
 * @apiVersion 1.0.0
 *
 * @apiDescription 获取系统当前的完整配置
 *
 * @apiSuccess {Object} config 配置对象
 * @apiSuccess {Object} config.server 服务端配置
 * @apiSuccess {Object} config.server.api API服务配置
 * @apiSuccess {String} config.server.api.address API服务监听地址，格式: "host:port" 或 ":port"
 * @apiSuccess {Object} config.server.ui Web界面配置
 * @apiSuccess {Boolean} config.server.ui.enabled 是否启用Web管理界面
 * @apiSuccess {String} config.server.ui.address Web界面监听地址
 * @apiSuccess {String} config.server.ui.path 静态文件目录路径
 * @apiSuccess {Object} [config.server.log] 日志配置
 * @apiSuccess {String} [config.server.log.level] 日志级别，可选值: "debug", "info", "warn", "error"
 * @apiSuccess {String} [config.server.log.file] 日志文件路径
 * @apiSuccess {Number} [config.server.log.maxSize] 单个日志文件最大大小，单位: MB
 * @apiSuccess {Number} [config.server.log.maxBackups] 保留的旧日志文件最大数量
 * @apiSuccess {Number} [config.server.log.maxAge] 保留旧日志文件的最大天数，单位: 天
 * @apiSuccess {Boolean} [config.server.log.compress] 是否压缩旧日志文件
 * @apiSuccess {Boolean} [config.server.log.enabled] 是否启用日志
 * @apiSuccess {Object[]} config.instances 代理实例配置列表
 * @apiSuccess {String} config.instances.name 实例名称，全局唯一标识符
 * @apiSuccess {String} config.instances.type 代理类型，可选值: "server", "client", "http-server", "http-client"
 * @apiSuccess {String} config.instances.listen 监听地址，格式: "host:port" 或 ":port"
 * @apiSuccess {String} config.instances.target 目标地址，格式: "host:port"
 * @apiSuccess {String} config.instances.protocol 协议类型，可选值: "auto", "tlcp", "tls"
 * @apiSuccess {Boolean} config.instances.enabled 是否启用该实例
 * @apiSuccess {Object} [config.instances.tlcp] TLCP协议专用配置
 * @apiSuccess {String} [config.instances.tlcp.auth] 认证模式，可选值: "none", "one-way", "mutual"
 * @apiSuccess {String} [config.instances.tlcp.minVersion] 最低协议版本，TLCP仅有"1.1"版本
 * @apiSuccess {String} [config.instances.tlcp.maxVersion] 最高协议版本，TLCP仅有"1.1"版本
 * @apiSuccess {String[]} [config.instances.tlcp.cipherSuites] 密码套件列表
 * @apiSuccess {String[]} [config.instances.tlcp.curvePreferences] 椭圆曲线偏好
 * @apiSuccess {Boolean} [config.instances.tlcp.sessionTickets] 是否启用会话票据
 * @apiSuccess {Boolean} [config.instances.tlcp.sessionCache] 是否启用会话缓存
 * @apiSuccess {Boolean} [config.instances.tlcp.insecureSkipVerify] 是否跳过证书验证（不安全，仅用于测试）
 * @apiSuccess {Object} [config.instances.tls] TLS协议专用配置
 * @apiSuccess {String} [config.instances.tls.auth] 认证模式，可选值: "none", "one-way", "mutual"
 * @apiSuccess {String} [config.instances.tls.minVersion] 最低协议版本，可选值: "1.0", "1.1", "1.2", "1.3"
 * @apiSuccess {String} [config.instances.tls.maxVersion] 最高协议版本，可选值: "1.0", "1.1", "1.2", "1.3"
 * @apiSuccess {String[]} [config.instances.tls.cipherSuites] 密码套件列表
 * @apiSuccess {String[]} [config.instances.tls.curvePreferences] 椭圆曲线偏好
 * @apiSuccess {Boolean} [config.instances.tls.sessionTickets] 是否启用会话票据
 * @apiSuccess {Boolean} [config.instances.tls.sessionCache] 是否启用会话缓存
 * @apiSuccess {Boolean} [config.instances.tls.insecureSkipVerify] 是否跳过证书验证（不安全，仅用于测试）
 * @apiSuccess {Object} [config.instances.certificates] 证书配置
 * @apiSuccess {Object} [config.instances.certificates.tlcp] TLCP协议证书配置
 * @apiSuccess {String} [config.instances.certificates.tlcp.cert] 证书文件路径，支持PEM格式
 * @apiSuccess {String} [config.instances.certificates.tlcp.key] 私钥文件路径，支持PEM格式
 * @apiSuccess {String} [config.instances.certificates.tlcp.keyName] 密钥存储名称
 * @apiSuccess {Object} [config.instances.certificates.tls] TLS协议证书配置
 * @apiSuccess {String} [config.instances.certificates.tls.cert] 证书文件路径，支持PEM格式
 * @apiSuccess {String} [config.instances.certificates.tls.key] 私钥文件路径，支持PEM格式
 * @apiSuccess {String} [config.instances.certificates.tls.keyName] 密钥存储名称
 * @apiSuccess {String[]} [config.instances.clientCa] 客户端CA证书路径列表
 * @apiSuccess {String[]} [config.instances.serverCa] 服务端CA证书路径列表
 * @apiSuccess {Object} [config.instances.http] HTTP协议专用配置
 * @apiSuccess {Object} [config.instances.http.requestHeaders] 请求头处理配置
 * @apiSuccess {Object} [config.instances.http.requestHeaders.add] 添加HTTP头
 * @apiSuccess {String[]} [config.instances.http.requestHeaders.remove] 删除指定的HTTP头
 * @apiSuccess {Object} [config.instances.http.requestHeaders.set] 设置HTTP头
 * @apiSuccess {Object} [config.instances.http.responseHeaders] 响应头处理配置
 * @apiSuccess {Object} [config.instances.http.responseHeaders.add] 添加HTTP头
 * @apiSuccess {String[]} [config.instances.http.responseHeaders.remove] 删除指定的HTTP头
 * @apiSuccess {Object} [config.instances.http.responseHeaders.set] 设置HTTP头
 * @apiSuccess {Object} [config.instances.log] 实例级别日志配置
 * @apiSuccess {Object} [config.instances.stats] 统计信息配置
 * @apiSuccess {Boolean} [config.instances.stats.enabled] 是否启用统计信息收集
 * @apiSuccess {Number} [config.instances.stats.interval] 统计信息收集间隔，单位: 纳秒
 * @apiSuccess {String} [config.instances.sni] 服务器名称指示
 * @apiSuccess {Object} [config.instances.timeout] 连接超时配置
 * @apiSuccess {Number} [config.instances.timeout.dial] 连接建立超时，默认: 10s
 * @apiSuccess {Number} [config.instances.timeout.read] 读取超时，默认: 30s
 * @apiSuccess {Number} [config.instances.timeout.write] 写入超时，默认: 30s
 * @apiSuccess {Number} [config.instances.timeout.handshake] TLS/TLCP握手超时，默认: 15s
 * @apiSuccess {String} [config.certDir] 证书文件目录路径
 * @apiSuccess {String} [config.trustedDir] 根证书目录路径
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "server": {
 *         "api": {
 *           "address": ":30080"
 *         },
 *         "ui": {
 *           "enabled": true,
 *           "address": ":30000",
 *           "path": "./ui"
 *         },
 *         "log": {
 *           "level": "info",
 *           "file": "./logs/tlcpchan.log",
 *           "maxSize": 100,
 *           "maxBackups": 5,
 *           "maxAge": 30,
 *           "compress": true,
 *           "enabled": true
 *         }
 *       },
 *       "instances": [
 *         {
 *           "name": "proxy-1",
 *           "type": "server",
 *           "listen": ":8443",
 *           "target": "backend:8080",
 *           "protocol": "auto",
 *           "enabled": true
 *         }
 *       ]
 *     }
 */
func (c *ConfigController) Get(w http.ResponseWriter, r *http.Request) {
	Success(w, c.cfg)
}

/**
 * @api {post} /api/v1/config 更新配置
 * @apiName UpdateConfig
 * @apiGroup Config
 * @apiVersion 1.0.0
 *
 * @apiDescription 更新系统配置并保存到文件
 *
 * @apiBody {Object} config 配置对象
 * @apiBody {Object} config.server 服务端配置
 * @apiBody {Object} config.server.api API服务配置
 * @apiBody {String} config.server.api.address API服务监听地址，格式: "host:port" 或 ":port"
 * @apiBody {Object} config.server.ui Web界面配置
 * @apiBody {Boolean} config.server.ui.enabled 是否启用Web管理界面
 * @apiBody {String} config.server.ui.address Web界面监听地址
 * @apiBody {String} config.server.ui.path 静态文件目录路径
 * @apiBody {Object} [config.server.log] 日志配置
 * @apiBody {String} [config.server.log.level] 日志级别，可选值: "debug", "info", "warn", "error"
 * @apiBody {String} [config.server.log.file] 日志文件路径
 * @apiBody {Number} [config.server.log.maxSize] 单个日志文件最大大小，单位: MB
 * @apiBody {Number} [config.server.log.maxBackups] 保留的旧日志文件最大数量
 * @apiBody {Number} [config.server.log.maxAge] 保留旧日志文件的最大天数，单位: 天
 * @apiBody {Boolean} [config.server.log.compress] 是否压缩旧日志文件
 * @apiBody {Boolean} [config.server.log.enabled] 是否启用日志
 * @apiBody {Object[]} config.instances 代理实例配置列表
 * @apiBody {String} config.instances.name 实例名称，全局唯一标识符
 * @apiBody {String} config.instances.type 代理类型，可选值: "server", "client", "http-server", "http-client"
 * @apiBody {String} config.instances.listen 监听地址，格式: "host:port" 或 ":port"
 * @apiBody {String} config.instances.target 目标地址，格式: "host:port"
 * @apiBody {String} config.instances.protocol 协议类型，可选值: "auto", "tlcp", "tls"
 * @apiBody {Boolean} config.instances.enabled 是否启用该实例
 * @apiBody {Object} [config.instances.tlcp] TLCP协议专用配置
 * @apiBody {String} [config.instances.tlcp.auth] 认证模式，可选值: "none", "one-way", "mutual"
 * @apiBody {String} [config.instances.tlcp.minVersion] 最低协议版本，TLCP仅有"1.1"版本
 * @apiBody {String} [config.instances.tlcp.maxVersion] 最高协议版本，TLCP仅有"1.1"版本
 * @apiBody {String[]} [config.instances.tlcp.cipherSuites] 密码套件列表
 * @apiBody {String[]} [config.instances.tlcp.curvePreferences] 椭圆曲线偏好
 * @apiBody {Boolean} [config.instances.tlcp.sessionTickets] 是否启用会话票据
 * @apiBody {Boolean} [config.instances.tlcp.sessionCache] 是否启用会话缓存
 * @apiBody {Boolean} [config.instances.tlcp.insecureSkipVerify] 是否跳过证书验证（不安全，仅用于测试）
 * @apiBody {Object} [config.instances.tls] TLS协议专用配置
 * @apiBody {String} [config.instances.tls.auth] 认证模式，可选值: "none", "one-way", "mutual"
 * @apiBody {String} [config.instances.tls.minVersion] 最低协议版本，可选值: "1.0", "1.1", "1.2", "1.3"
 * @apiBody {String} [config.instances.tls.maxVersion] 最高协议版本，可选值: "1.0", "1.1", "1.2", "1.3"
 * @apiBody {String[]} [config.instances.tls.cipherSuites] 密码套件列表
 * @apiBody {String[]} [config.instances.tls.curvePreferences] 椭圆曲线偏好
 * @apiBody {Boolean} [config.instances.tls.sessionTickets] 是否启用会话票据
 * @apiBody {Boolean} [config.instances.tls.sessionCache] 是否启用会话缓存
 * @apiBody {Boolean} [config.instances.tls.insecureSkipVerify] 是否跳过证书验证（不安全，仅用于测试）
 * @apiBody {Object} [config.instances.certificates] 证书配置
 * @apiBody {Object} [config.instances.certificates.tlcp] TLCP协议证书配置
 * @apiBody {String} [config.instances.certificates.tlcp.cert] 证书文件路径，支持PEM格式
 * @apiBody {String} [config.instances.certificates.tlcp.key] 私钥文件路径，支持PEM格式
 * @apiBody {String} [config.instances.certificates.tlcp.keyName] 密钥存储名称
 * @apiBody {Object} [config.instances.certificates.tls] TLS协议证书配置
 * @apiBody {String} [config.instances.certificates.tls.cert] 证书文件路径，支持PEM格式
 * @apiBody {String} [config.instances.certificates.tls.key] 私钥文件路径，支持PEM格式
 * @apiBody {String} [config.instances.certificates.tls.keyName] 密钥存储名称
 * @apiBody {String[]} [config.instances.clientCa] 客户端CA证书路径列表
 * @apiBody {String[]} [config.instances.serverCa] 服务端CA证书路径列表
 * @apiBody {Object} [config.instances.http] HTTP协议专用配置
 * @apiBody {Object} [config.instances.http.requestHeaders] 请求头处理配置
 * @apiBody {Object} [config.instances.http.requestHeaders.add] 添加HTTP头
 * @apiBody {String[]} [config.instances.http.requestHeaders.remove] 删除指定的HTTP头
 * @apiBody {Object} [config.instances.http.requestHeaders.set] 设置HTTP头
 * @apiBody {Object} [config.instances.http.responseHeaders] 响应头处理配置
 * @apiBody {Object} [config.instances.http.responseHeaders.add] 添加HTTP头
 * @apiBody {String[]} [config.instances.http.responseHeaders.remove] 删除指定的HTTP头
 * @apiBody {Object} [config.instances.http.responseHeaders.set] 设置HTTP头
 * @apiBody {Object} [config.instances.log] 实例级别日志配置
 * @apiBody {Object} [config.instances.stats] 统计信息配置
 * @apiBody {Boolean} [config.instances.stats.enabled] 是否启用统计信息收集
 * @apiBody {Number} [config.instances.stats.interval] 统计信息收集间隔，单位: 纳秒
 * @apiBody {String} [config.instances.sni] 服务器名称指示
 * @apiBody {Object} [config.instances.timeout] 连接超时配置
 * @apiBody {Number} [config.instances.timeout.dial] 连接建立超时，默认: 10s
 * @apiBody {Number} [config.instances.timeout.read] 读取超时，默认: 30s
 * @apiBody {Number} [config.instances.timeout.write] 写入超时，默认: 30s
 * @apiBody {Number} [config.instances.timeout.handshake] TLS/TLCP握手超时，默认: 15s
 * @apiBody {String} [config.certDir] 证书文件目录路径
 * @apiBody {String} [config.trustedDir] 根证书目录路径
 *
 * @apiParamExample {json} Request-Example:
 *     {
 *       "server": {
 *         "api": {
 *           "address": ":30080"
 *         },
 *         "ui": {
 *           "enabled": true,
 *           "address": ":30000",
 *           "path": "./ui"
 *         },
 *         "log": {
 *           "level": "info",
 *           "file": "./logs/tlcpchan.log",
 *           "maxSize": 100,
 *           "maxBackups": 5,
 *           "maxAge": 30,
 *           "compress": true,
 *           "enabled": true
 *         }
 *       },
 *       "instances": [
 *         {
 *           "name": "proxy-1",
 *           "type": "server",
 *           "listen": ":8443",
 *           "target": "backend:8080",
 *           "protocol": "auto",
 *           "enabled": true
 *         }
 *       ]
 *     }
 *
 * @apiSuccess {Object} config 更新后的配置对象
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "server": {
 *         "api": {
 *           "address": ":30080"
 *         },
 *         "ui": {
 *           "enabled": true,
 *           "address": ":30000",
 *           "path": "./ui"
 *         },
 *         "log": {
 *           "level": "info",
 *           "file": "./logs/tlcpchan.log",
 *           "maxSize": 100,
 *           "maxBackups": 5,
 *           "maxAge": 30,
 *           "compress": true,
 *           "enabled": true
 *         }
 *       },
 *       "instances": [
 *         {
 *           "name": "proxy-1",
 *           "type": "server",
 *           "listen": ":8443",
 *           "target": "backend:8080",
 *           "protocol": "auto",
 *           "enabled": true
 *         }
 *       ]
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     无效的请求体
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     配置验证失败
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     保存配置失败
 */
func (c *ConfigController) Update(w http.ResponseWriter, r *http.Request) {
	var newCfg config.Config
	if err := parseJSON(r, &newCfg); err != nil {
		BadRequest(w, "无效的请求体: "+err.Error())
		return
	}

	if err := config.Validate(&newCfg); err != nil {
		BadRequest(w, "配置验证失败: "+err.Error())
		return
	}

	if err := config.Save(c.configPath, &newCfg); err != nil {
		InternalError(w, "保存配置失败: "+err.Error())
		return
	}

	c.cfg = &newCfg
	c.log.Info("配置已更新")
	Success(w, c.cfg)
}

/**
 * @api {post} /api/v1/config/reload 重载配置
 * @apiName ReloadConfig
 * @apiGroup Config
 * @apiVersion 1.0.0
 *
 * @apiDescription 从配置文件重新加载系统配置
 *
 * @apiSuccess {Object} config 重新加载后的配置对象
 * @apiSuccess {Object} config.server 服务端配置
 * @apiSuccess {Object} config.server.api API服务配置
 * @apiSuccess {String} config.server.api.address API服务监听地址
 * @apiSuccess {Object} config.server.ui Web界面配置
 * @apiSuccess {Boolean} config.server.ui.enabled 是否启用Web管理界面
 * @apiSuccess {String} config.server.ui.address Web界面监听地址
 * @apiSuccess {String} config.server.ui.path 静态文件目录路径
 * @apiSuccess {Object} [config.server.log] 日志配置
 * @apiSuccess {Object[]} config.instances 代理实例配置列表
 * @apiSuccess {String} config.instances.name 实例名称
 * @apiSuccess {String} config.instances.type 代理类型
 * @apiSuccess {String} config.instances.listen 监听地址
 * @apiSuccess {String} config.instances.target 目标地址
 * @apiSuccess {String} config.instances.protocol 协议类型
 * @apiSuccess {Boolean} config.instances.enabled 是否启用该实例
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "server": {
 *         "api": {
 *           "address": ":30080"
 *         },
 *         "ui": {
 *           "enabled": true,
 *           "address": ":30000",
 *           "path": "./ui"
 *         },
 *         "log": {
 *           "level": "info",
 *           "file": "./logs/tlcpchan.log",
 *           "maxSize": 100,
 *           "maxBackups": 5,
 *           "maxAge": 30,
 *           "compress": true,
 *           "enabled": true
 *         }
 *       },
 *       "instances": [
 *         {
 *           "name": "proxy-1",
 *           "type": "server",
 *           "listen": ":8443",
 *           "target": "backend:8080",
 *           "protocol": "auto",
 *           "enabled": true
 *         }
 *       ]
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     重新加载配置失败
 */
func (c *ConfigController) Reload(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load(c.configPath)
	if err != nil {
		InternalError(w, "重新加载配置失败: "+err.Error())
		return
	}

	c.cfg = cfg
	c.log.Info("配置已重新加载")
	Success(w, c.cfg)
}

func (c *ConfigController) RegisterRoutes(router *Router) {
	router.GET("/api/v1/config", c.Get)
	router.POST("/api/v1/config", c.Update)
	router.POST("/api/v1/config/reload", c.Reload)
}
