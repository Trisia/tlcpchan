package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/instance"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/proxy"
)

type InstanceController struct {
	manager    *instance.Manager
	configPath string
	log        *logger.Logger
}

func NewInstanceController(mgr *instance.Manager, configPath string) *InstanceController {
	return &InstanceController{
		manager:    mgr,
		configPath: configPath,
		log:        logger.Default(),
	}
}

/**
 * @api {get} /api/instances 获取实例列表
 * @apiName ListInstances
 * @apiGroup Instance
 * @apiVersion 1.0.0
 *
 * @apiDescription 获取所有代理实例的列表信息
 *
 * @apiSuccess {Object[]} - 实例列表数组
 * @apiSuccess {String} -.name 实例名称，唯一标识符
 * @apiSuccess {String} -.status 实例状态，可选值：created（已创建）、running（运行中）、stopped（已停止）、error（错误）
 * @apiSuccess {Object} -.config 实例配置对象
 * @apiSuccess {Boolean} -.enabled 是否启用，true 表示启用，false 表示禁用
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     [
 *       {
 *         "name": "tlcp-server",
 *         "status": "running",
 *         "config": {
 *           "name": "tlcp-server",
 *           "type": "server",
 *           "protocol": "tlcp",
 *           "listen": ":443",
 *           "target": "127.0.0.1:8080",
 *           "enabled": true
 *         },
 *         "enabled": true
 *       },
 *       {
 *         "name": "tls-client",
 *         "status": "stopped",
 *         "config": {
 *           "name": "tls-client",
 *           "type": "client",
 *           "protocol": "tls",
 *           "listen": ":8443",
 *           "target": "backend.example.com:443",
 *           "enabled": false
 *         },
 *         "enabled": false
 *       }
 *     ]
 */
func (c *InstanceController) List(w http.ResponseWriter, r *http.Request) {
	instances := c.manager.List()
	data := make([]map[string]interface{}, len(instances))
	for i, inst := range instances {
		data[i] = map[string]interface{}{
			"name":    inst.Name(),
			"status":  inst.Status(),
			"config":  inst.Config(),
			"enabled": inst.Config().Enabled,
		}
	}
	Success(w, data)
}

/**
 * @api {get} /api/instances/:name 获取实例详情
 * @apiName GetInstance
 * @apiGroup Instance
 * @apiVersion 1.0.0
 *
 * @apiDescription 获取指定实例的详细信息，包括配置和状态
 *
 * @apiParam {String} name 实例名称（路径参数），实例的唯一标识符
 *
 * @apiSuccess {String} name 实例名称
 * @apiSuccess {String} status 实例状态，可选值：created（已创建）、running（运行中）、stopped（已停止）、error（错误）
 * @apiSuccess {Object} config 实例完整配置对象
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "name": "tlcp-server",
 *       "status": "running",
 *       "config": {
 *         "name": "tlcp-server",
 *         "type": "server",
 *         "protocol": "tlcp",
 *         "auth": "mutual",
 *         "listen": ":443",
 *         "target": "127.0.0.1:8080",
 *         "enabled": true,
 *         "certificates": {
 *           "tlcp": {
 *             "cert": "server-sm2",
 *             "key": "server-sm2"
 *           }
 *         },
 *         "client_ca": ["ca-sm2"],
 *         "tlcp": {
 *           "min_version": "1.1",
 *           "max_version": "1.1",
 *           "cipher_suites": ["ECC_SM4_GCM_SM3", "ECDHE_SM4_GCM_SM3"],
 *           "session_tickets": true
 *         }
 *       }
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     Content-Type: text/plain
 *
 *     实例不存在
 */
func (c *InstanceController) Get(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	inst, ok := c.manager.Get(name)
	if !ok {
		NotFound(w, "实例不存在")
		return
	}
	Success(w, map[string]interface{}{
		"name":   inst.Name(),
		"status": inst.Status(),
		"config": inst.Config(),
	})
}

/**
 * @api {post} /api/instances 创建实例
 * @apiName CreateInstance
 * @apiGroup Instance
 * @apiVersion 1.0.0
 *
 * @apiDescription 创建一个新的代理实例
 *
 * @apiBody {String} name 实例名称，唯一标识符，只能包含字母、数字、下划线和连字符
 * @apiBody {String} type 实例类型，可选值：server（服务端）、client（客户端）、http-server（HTTP服务端）、http-client（HTTP客户端）
 * @apiBody {String} protocol 协议类型，可选值：auto（自动检测）、tlcp（仅TLCP）、tls（仅TLS）
 * @apiBody {String} [auth=none] 认证模式，可选值：none（无认证）、one-way（单向认证）、mutual（双向认证）
 * @apiBody {String} listen 监听地址，格式为 ":port" 或 "ip:port"，例如 ":443" 或 "127.0.0.1:8443"
 * @apiBody {String} target 目标地址，格式为 "host:port"，例如 "backend.example.com:8080"
 * @apiBody {Boolean} [enabled=true] 是否启用，true 表示创建后自动启动，false 表示创建后保持停止状态
 * @apiBody {Object} [certificates] 证书配置对象
 * @apiBody {Object} [certificates.tlcp] TLCP证书配置
 * @apiBody {String} [certificates.tlcp.cert] TLCP证书名称，对应证书目录中的证书文件名
 * @apiBody {String} [certificates.tlcp.key] TLCP私钥名称，对应证书目录中的私钥文件名
 * @apiBody {Object} [certificates.tls] TLS证书配置
 * @apiBody {String} [certificates.tls.cert] TLS证书名称
 * @apiBody {String} [certificates.tls.key] TLS私钥名称
 * @apiBody {String[]} [client_ca] 客户端CA证书名称列表，用于双向认证时验证客户端证书
 * @apiBody {Object} [tlcp] TLCP协议配置
 * @apiBody {String} [tlcp.client_auth_type=no-client-cert] TLCP客户端认证类型，可选值："no-client-cert"、"request-client-cert"、"require-any-client-cert"、"verify-client-cert-if-given"、"require-and-verify-client-cert"
 * @apiBody {String} [tlcp.min_version=1.1] TLCP最小版本，可选值："1.1"
 * @apiBody {String} [tlcp.max_version=1.1] TLCP最大版本，可选值："1.1"
 * @apiBody {String[]} [tlcp.cipher_suites] TLCP密码套件列表，可选值："ECC_SM4_CBC_SM3"、"ECC_SM4_GCM_SM3"、"ECDHE_SM4_CBC_SM3"、"ECDHE_SM4_GCM_SM3" 等
 * @apiBody {Boolean} [tlcp.session_tickets=true] 是否启用会话票证
 * @apiBody {Object} [tls] TLS协议配置
 * @apiBody {String} [tls.client_auth_type=no-client-cert] TLS客户端认证类型，可选值："no-client-cert"、"request-client-cert"、"require-any-client-cert"、"verify-client-cert-if-given"、"require-and-verify-client-cert"
 * @apiBody {String} [tls.min_version=1.2] TLS最小版本，可选值："1.0"、"1.1"、"1.2"、"1.3"
 * @apiBody {String} [tls.max_version=1.3] TLS最大版本，可选值："1.0"、"1.1"、"1.2"、"1.3"
 * @apiBody {String[]} [tls.cipher_suites] TLS密码套件列表
 * @apiBody {String} [sni] SNI（Server Name Indication）服务器名称指示，用于TLS客户端连接时指定服务器名称
 *
 * @apiSuccess {String} name 实例名称
 * @apiSuccess {String} status 实例状态，创建后通常为 "created" 或 "running"
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 201 Created
 *     {
 *       "name": "tlcp-server",
 *       "status": "created"
 *     }
 *
 * @apiParamExample {json} Request-Example:
 *     {
 *       "name": "tlcp-server",
 *       "type": "server",
 *       "protocol": "tlcp",
 *       "auth": "mutual",
 *       "listen": ":443",
 *       "target": "127.0.0.1:8080",
 *       "enabled": true,
 *       "certificates": {
 *         "tlcp": {
 *           "cert": "server-sm2",
 *           "key": "server-sm2"
 *         }
 *       },
 *       "client_ca": ["ca-sm2"],
 *       "tlcp": {
 *         "min_version": "1.1",
 *         "max_version": "1.1",
 *         "cipher_suites": ["ECC_SM4_GCM_SM3", "ECDHE_SM4_GCM_SM3"],
 *         "session_tickets": true
 *       }
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     Content-Type: text/plain
 *
 *     无效的请求体: json: cannot unmarshal string into Go value of type config.InstanceConfig
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     Content-Type: text/plain
 *
 *     实例名称不能为空
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 409 Conflict
 *     Content-Type: text/plain
 *
 *     实例 tlcp-server 已存在
 */
func (c *InstanceController) Create(w http.ResponseWriter, r *http.Request) {
	var cfg config.InstanceConfig
	if err := parseJSON(r, &cfg); err != nil {
		BadRequest(w, "无效的请求体: "+err.Error())
		return
	}

	inst, err := c.manager.Create(&cfg)
	if err != nil {
		BadRequest(w, err.Error())
		return
	}

	c.log.Info("创建实例: %s", cfg.Name)
	Created(w, map[string]interface{}{
		"name":   inst.Name(),
		"status": inst.Status(),
	})
}

/**
 * @api {put} /api/instances/:name 更新实例配置
 * @apiName UpdateInstance
 * @apiGroup Instance
 * @apiVersion 1.0.0
 *
 * @apiDescription 更新正在运行的实例配置（热更新），只能更新运行中的实例
 *
 * @apiParam {String} name 实例名称（路径参数），实例的唯一标识符
 *
 * @apiBody {String} [type] 实例类型，可选值：server、client、http-server、http-client
 * @apiBody {String} [protocol] 协议类型，可选值：auto、tlcp、tls
 * @apiBody {String} [auth] 认证模式，可选值：none、one-way、mutual
 * @apiBody {String} [listen] 监听地址，格式为 ":port" 或 "ip:port"
 * @apiBody {String} [target] 目标地址，格式为 "host:port"
 * @apiBody {Boolean} [enabled] 是否启用
 * @apiBody {Object} [certificates] 证书配置对象
 * @apiBody {String[]} [client_ca] 客户端CA证书名称列表
 * @apiBody {Object} [tlcp] TLCP协议配置
 * @apiBody {Object} [tls] TLS协议配置
 *
 * @apiSuccess {String} name 实例名称
 * @apiSuccess {String} status 实例状态，更新后通常保持 "running"
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "name": "tlcp-server",
 *       "status": "running"
 *     }
 *
 * @apiParamExample {json} Request-Example:
 *     {
 *       "target": "127.0.0.1:9090",
 *       "enabled": true
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     Content-Type: text/plain
 *
 *     实例不存在
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     Content-Type: text/plain
 *
 *     实例未运行，无法热更新
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     Content-Type: text/plain
 *
 *     无效的请求体
 */
func (c *InstanceController) Update(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	inst, ok := c.manager.Get(name)
	if !ok {
		NotFound(w, "实例不存在")
		return
	}

	var cfg config.InstanceConfig
	if err := parseJSON(r, &cfg); err != nil {
		BadRequest(w, "无效的请求体: "+err.Error())
		return
	}

	cfg.Name = name
	if inst.Status() == instance.StatusRunning {
		if err := inst.Reload(&cfg); err != nil {
			BadRequest(w, err.Error())
			return
		}
	} else {
		BadRequest(w, "实例未运行，无法热更新")
		return
	}

	c.log.Info("更新实例配置: %s", name)
	Success(w, map[string]interface{}{
		"name":   inst.Name(),
		"status": inst.Status(),
	})
}

/**
 * @api {delete} /api/instances/:name 删除实例
 * @apiName DeleteInstance
 * @apiGroup Instance
 * @apiVersion 1.0.0
 *
 * @apiDescription 删除指定的代理实例，删除前会自动停止实例
 *
 * @apiParam {String} name 实例名称（路径参数），实例的唯一标识符
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     null
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     Content-Type: text/plain
 *
 *     删除实例失败：实例正在使用中
 */
func (c *InstanceController) Delete(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	if err := c.manager.Delete(name); err != nil {
		BadRequest(w, err.Error())
		return
	}
	c.log.Info("删除实例: %s", name)
	Success(w, nil)
}

/**
 * @api {post} /api/instances/:name/start 启动实例
 * @apiName StartInstance
 * @apiGroup Instance
 * @apiVersion 1.0.0
 *
 * @apiDescription 启动指定的代理实例，实例将开始监听端口并转发流量
 *
 * @apiParam {String} name 实例名称（路径参数），实例的唯一标识符
 *
 * @apiSuccess {String} status 实例状态，启动成功后为 "running"
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "status": "running"
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     Content-Type: text/plain
 *
 *     实例不存在
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     Content-Type: text/plain
 *
 *     启动实例失败：端口已被占用
 */
func (c *InstanceController) Start(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	inst, ok := c.manager.Get(name)
	if !ok {
		NotFound(w, "实例不存在")
		return
	}

	if err := inst.Start(); err != nil {
		BadRequest(w, err.Error())
		return
	}

	c.log.Info("启动实例: %s", name)
	Success(w, map[string]string{"status": string(inst.Status())})
}

/**
 * @api {post} /api/instances/:name/stop 停止实例
 * @apiName StopInstance
 * @apiGroup Instance
 * @apiVersion 1.0.0
 *
 * @apiDescription 停止指定的代理实例，实例将停止监听端口，现有连接会被优雅关闭
 *
 * @apiParam {String} name 实例名称（路径参数），实例的唯一标识符
 *
 * @apiSuccess {String} status 实例状态，停止成功后为 "stopped"
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "status": "stopped"
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     Content-Type: text/plain
 *
 *     实例不存在
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     Content-Type: text/plain
 *
 *     停止实例失败
 */
func (c *InstanceController) Stop(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	inst, ok := c.manager.Get(name)
	if !ok {
		NotFound(w, "实例不存在")
		return
	}

	if err := inst.Stop(); err != nil {
		BadRequest(w, err.Error())
		return
	}

	c.log.Info("停止实例: %s", name)
	Success(w, map[string]string{"status": string(inst.Status())})
}

/**
 * @api {post} /api/instances/:name/reload 热重载实例
 * @apiName ReloadInstance
 * @apiGroup Instance
 * @apiVersion 1.0.0
 *
 * @apiDescription 热重载指定实例的配置，不中断服务
 *
 * @apiParam {String} name 实例名称（路径参数），实例的唯一标识符
 *
 * @apiSuccess {String} status 实例状态，重载后通常保持 "running"
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "status": "running"
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     Content-Type: text/plain
 *
 *     实例不存在
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     Content-Type: text/plain
 *
 *     热重载实例失败
 */
func (c *InstanceController) Reload(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	inst, ok := c.manager.Get(name)
	if !ok {
		NotFound(w, "实例不存在")
		return
	}

	cfg := inst.Config()
	if err := inst.Reload(cfg); err != nil {
		BadRequest(w, err.Error())
		return
	}

	c.log.Info("热重载实例: %s", name)
	Success(w, map[string]string{"status": string(inst.Status())})
}

/**
 * @api {post} /api/instances/:name/restart 重启实例
 * @apiName RestartInstance
 * @apiGroup Instance
 * @apiVersion 1.0.0
 *
 * @apiDescription 重启指定实例，停止后使用当前配置重新启动
 *
 * @apiParam {String} name 实例名称（路径参数），实例的唯一标识符
 *
 * @apiSuccess {String} status 实例状态，重启后为 "running"
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "status": "running"
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     Content-Type: text/plain
 *
 *     实例不存在
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     Content-Type: text/plain
 *
 *     重启实例失败
 */
func (c *InstanceController) Restart(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	inst, ok := c.manager.Get(name)
	if !ok {
		NotFound(w, "实例不存在")
		return
	}

	cfg := inst.Config()
	if err := inst.Restart(cfg); err != nil {
		BadRequest(w, err.Error())
		return
	}

	c.log.Info("重启实例: %s", name)
	Success(w, map[string]string{"status": string(inst.Status())})
}

/**
 * @api {get} /api/instances/:name/stats 获取实例统计
 * @apiName GetInstanceStats
 * @apiGroup Instance
 * @apiVersion 1.0.0
 *
 * @apiDescription 获取指定实例的运行统计信息，包括连接数、流量等
 *
 * @apiParam {String} name 实例名称（路径参数），实例的唯一标识符
 * @apiQuery {String} [period] 统计周期，可选值：1m（1分钟）、5m（5分钟）、15m（15分钟）、1h（1小时）、1d（1天）
 *
 * @apiSuccess {Number} totalConnections 总连接数，自实例启动以来的累计连接数
 * @apiSuccess {Number} activeConnections 当前活跃连接数，当前正在处理的连接数
 * @apiSuccess {Number} bytesReceived 接收字节数，单位：字节，自实例启动以来累计接收的数据量
 * @apiSuccess {Number} bytesSent 发送字节数，单位：字节，自实例启动以来累计发送的数据量
 * @apiSuccess {Number} [requestsTotal] 总请求数，HTTP模式下的请求总数
 * @apiSuccess {Number} [errors] 错误数，发生的错误总数
 * @apiSuccess {Number} [avgLatencyMs] 平均延迟，单位：毫秒，请求处理的平均延迟
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "totalConnections": 1000,
 *       "activeConnections": 10,
 *       "bytesReceived": 1048576,
 *       "bytesSent": 2097152,
 *       "requestsTotal": 500,
 *       "errors": 2,
 *       "avgLatencyMs": 5.2
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     Content-Type: text/plain
 *
 *     实例不存在
 */
func (c *InstanceController) Stats(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	inst, ok := c.manager.Get(name)
	if !ok {
		NotFound(w, "实例不存在")
		return
	}

	Success(w, inst.Stats())
}

/**
 * @api {get} /api/instances/:name/logs 获取实例日志
 * @apiName GetInstanceLogs
 * @apiGroup Instance
 * @apiVersion 1.0.0
 *
 * @apiDescription 获取指定实例的日志信息
 *
 * @apiParam {String} name 实例名称（路径参数），实例的唯一标识符
 * @apiQuery {Number} [lines=100] 返回日志行数，默认值：100，最大值：1000
 * @apiQuery {String} [level] 日志级别过滤，可选值：debug（调试）、info（信息）、warn（警告）、error（错误）
 *
 * @apiSuccess {Object[]} logs 日志列表数组
 * @apiSuccess {String} logs.timestamp 时间戳，ISO 8601 格式，例如 "2024-01-01T00:00:00Z"
 * @apiSuccess {String} logs.level 日志级别，可选值：debug、info、warn、error
 * @apiSuccess {String} logs.message 日志消息内容
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "logs": [
 *         {
 *           "timestamp": "2024-01-01T00:00:00Z",
 *           "level": "info",
 *           "message": "connection accepted from 192.168.1.1:12345"
 *         },
 *         {
 *           "timestamp": "2024-01-01T00:00:01Z",
 *           "level": "warn",
 *           "message": "client certificate validation failed"
 *         }
 *       ]
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     Content-Type: text/plain
 *
 *     实例不存在
 */
func (c *InstanceController) Logs(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	_, ok := c.manager.Get(name)
	if !ok {
		NotFound(w, "实例不存在")
		return
	}

	Success(w, []map[string]interface{}{
		{"timestamp": "2024-01-01T00:00:00Z", "level": "info", "message": "示例日志"},
	})
}

/**
 * @api {put} /api/instances/:name/edit 编辑实例配置
 * @apiName EditInstance
 * @apiGroup Instance
 * @apiVersion 1.0.0
 *
 * * @apiDescription 编辑实例配置，保存到配置文件并在实例运行时热重载
 *
 * @apiParam {String} name 实例名称（路径参数），实例的唯一标识符
 *
 * @apiBody {String} type 实例类型，可选值：server（服务端）、client（客户端）、http-server（HTTP服务端）、http-client（HTTP客户端）
 * @apiBody {String} protocol 协议类型，可选值：auto（自动检测）、tlcp（仅TLCP）、tls（仅TLS）
 * * @apiBody {String} [auth=none] 认证模式，可选值：none（无认证）、one-way（单向认证）、mutual（双向认证）
 * @apiBody {String} listen 监听地址，格式为 ":port" 或 "ip:port"，例如 ":443" 或 "127.0.0.1:8443"
 * @apiBody {String} target 目标地址，格式为 "host:port"，例如 "backend.example.com:8080"
 * @apiBody {Boolean} enabled 是否启用，true 表示启用，false 表示禁用
 * @apiBody {Object} [certificates] 证书配置对象
 * @apiBody {Object} [certificates.tlcp] TLCP证书配置
 * @apiBody {String} [certificates.tlcp.cert] TLCP证书名称，对应证书目录中的证书文件名
 * @apiBody {String} [certificates.tlcp.key] TLCP私钥名称，对应证书目录中的私钥文件名
 * @apiBody {Object} [certificates.tls] TLS证书配置
 * @apiBody {String} [certificates.tls.cert] TLS证书名称
 * @apiBody {String} [certificates.tls.key] TLS私钥名称
 * @apiBody {String[]} [client_ca] 客户端CA证书名称列表，用于双向认证时验证客户端证书
 * @apiBody {String[]} [server_ca] 服务端CA证书名称列表，用于验证服务端证书
 * @apiBody {Object} [tlcp] TLCP协议配置
 * @apiBody {String} [tlcp.min_version=1.1] TLCP最小版本，可选值："1.1"
 * @apiBody {String} [tlcp.max_version=1.1] TLCP最大版本，可选值："1.1"
 * @apiBody {String[]} [tlcp.cipher_suites] TLCP密码套件列表，可选值："ECC_SM4_CBC_SM3"、"ECC_SM4_GCM_SM3"、"ECDHE_SM4_CBC_SM3"、"ECDHE_SM4_GCM_SM3" 等
 * @apiBody {Boolean} [tlcp.session_tickets=true] 是否启用会话票证
 * @apiBody {Object} [tls] TLS协议配置
 * @apiBody {String} [tls.min_version=1.2] TLS最小版本，可选值："1.0"、"1.1"、"1.2"、"1.3"
 * @apiBody {String} [tls.max_version=1.3] TLS最大版本，可选值："1.0"、"1.1"、"1.2"、"1.3"
 * @apiBody {String[]} [tls.cipher_suites] TLS密码套件列表
 * @apiBody {Object} [http] HTTP协议配置
 * @apiBody {Boolean} [http.compression] 是否启用压缩
 * @apiBody {String} [http.compressionLevel] 压缩级别
 * @apiBody {String} sni SNI（Server Name Indication）服务器名称指示，用于TLS客户端连接时指定服务器名称
 * @apiBody {Object} [timeout] 超时配置
 * @apiBody {Number} [timeout.read] 读超时时间，单位：秒
 * @apiBody {Number} [timeout.write] 写超时时间，单位：秒
 * @apiBody {Number} [timeout.handshake] 握手超时时间，单位：秒
 * @apiBody {Number} [bufferSize=4096] 缓冲区大小，单位：字节，必须是正整数
 *
 * @apiSuccess {String} name 实例名称
 * @apiSuccess {String} status 实例状态，可选值：created（已创建）、running（运行中）、stopped（已停止）、error（错误）
 * @apiSuccess {Object} config 更新后的完整配置对象
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "name": "tlcp-server",
 *       "status": "running",
 *       "config": {
 *         "name": "tlcp-server",
 *         "type": "server",
 *         "protocol": "tlcp",
 *         "auth": "mutual",
 *         "listen": ":443",
 *         "target": "127.0.0.1:8080",
 *         "enabled": true,
 *         "certificates": {
 *           "tlcp": {
 *             "cert": "server-sm2",
 *             "key": "server-sm2"
 *           }
 *         },
 *         "client_ca": ["ca-sm2"],
 *         "tlcp": {
 *           "min_version": "1.1",
 *           "max_version": "1.1",
 *           "cipher_suites": ["ECC_SM4_GCM_SM3", "ECDHE_SM4_GCM_SM3"],
 *           "session_tickets": true
 *         }
 *       }
 *     }
 *
 * @apiParamExample {json} Request-Example:
 *     {
 *       "type": "server",
 *       "protocol": "tlcp",
 *       "auth": "mutual",
 *       "listen": ":443",
 *       "target": "127.0.0.1:8080",
 *       "enabled": true,
 *       "certificates": {
 *         "tlcp": {
 *           "cert": "server-sm2",
 *           "key": "server-sm2"
 *         }
 *       },
 *       "client_ca": ["ca-sm2"]
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     Content-Type: text/plain
 *
 *     无效的请求体: json: cannot unmarshal string into Go value of type config.InstanceConfig
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     Content-Type: text/plain
 *
 *     实例不存在
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     Content-Type: text/plain
 *
 *     配置验证失败: listen address is required
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     Content-Type: text/plain
 *
 *     热重载失败: port already in use
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     Content-Type: text/plain
 *
 *     保存配置失败: permission denied
 */
func (c *InstanceController) Edit(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")

	inst, ok := c.manager.Get(name)
	if !ok {
		NotFound(w, "实例不存在")
		return
	}

	var newCfg config.InstanceConfig
	if err := parseJSON(r, &newCfg); err != nil {
		BadRequest(w, "无效的请求体: "+err.Error())
		return
	}

	newCfg.Name = name

	currentCfg := config.Get()
	found := false
	for i, instance := range currentCfg.Instances {
		if instance.Name == name {
			currentCfg.Instances[i] = newCfg
			found = true
			break
		}
	}
	if !found {
		NotFound(w, "实例不存在")
		return
	}

	if err := config.Save(c.configPath, currentCfg); err != nil {
		InternalError(w, "保存配置失败: "+err.Error())
		return
	}
	config.Set(currentCfg)

	c.log.Info("实例配置已保存: %s", name)

	if inst.Status() == instance.StatusRunning {
		if err := inst.Reload(&newCfg); err != nil {
			BadRequest(w, "热重载失败: "+err.Error())
			return
		}
		c.log.Info("实例已热重载: %s", name)
	}

	Success(w, map[string]interface{}{
		"name":   name,
		"status": inst.Status(),
		"config": newCfg,
	})
}

/**
 * @api {get} /api/instances/:name/health 实例健康检查
 * @apiName InstanceHealthCheck
 * @apiGroup Instance
 * @apiVersion 1.0.0
 *
 * @apiDescription 检查代理实例的健康状态，通过建立连接测试配置的目标地址
 *
 * @apiParam {String} name 实例名称
 * @apiParam {Number} [timeout] 超时时间（秒），默认 10 秒
 * @apiParam {String} [protocol] 协议类型，可选值: tlcp, tls。auto 模式下使用此参数指定测试协议
 *
 * @apiSuccess {String} instance 实例名称
 * @apiSuccess {Object[]} results 健康检查结果数组
 * @apiSuccess {String} results.protocol 协议类型
 * @apiSuccess {Boolean} results.success 是否成功
 * @apiSuccess {Number} results.latencyMs 延迟（毫秒）
 * @apiSuccess {String} [results.error] 错误信息
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "instance": "proxy-1",
 *       "results": [
 *         {
 *           "protocol": "tlcp",
 *           "success": true,
 *           "latencyMs": 123
 *         },
 *         {
 *           "protocol": "tls",
 *           "success": false,
 *           "latencyMs": 0,
 *           "error": "连接超时"
 *         }
 *       ]
 *     }
 */
func (c *InstanceController) InstanceHealth(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	if name == "" {
		BadRequest(w, "实例名称不能为空")
		return
	}

	inst, ok := c.manager.Get(name)
	if !ok {
		NotFound(w, "实例不存在")
		return
	}

	timeout := 10 * time.Second
	if timeoutStr := r.URL.Query().Get("timeout"); timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil && t > 0 {
			timeout = time.Duration(t) * time.Second
		}
	}

	protocolParam := r.URL.Query().Get("protocol")
	instanceProtocol := inst.Protocol()

	var results []*proxy.HealthCheckResult

	if instanceProtocol == string(config.ProtocolAuto) {
		if protocolParam != "" {
			p := proxy.ParseProtocolType(protocolParam)
			if p == proxy.ProtocolTLCP || p == proxy.ProtocolTLS {
				result := inst.CheckHealth(p, timeout)
				results = append(results, result)
			} else {
				BadRequest(w, "无效的协议类型，可选值: tlcp, tls")
				return
			}
		} else {
			tlcpResult := inst.CheckHealth(proxy.ProtocolTLCP, timeout)
			tlsResult := inst.CheckHealth(proxy.ProtocolTLS, timeout)
			results = append(results, tlcpResult, tlsResult)
		}
	} else if instanceProtocol == string(config.ProtocolTLCP) {
		result := inst.CheckHealth(proxy.ProtocolTLCP, timeout)
		results = append(results, result)
	} else if instanceProtocol == string(config.ProtocolTLS) {
		result := inst.CheckHealth(proxy.ProtocolTLS, timeout)
		results = append(results, result)
	}

	Success(w, map[string]interface{}{
		"instance": name,
		"results":  results,
	})
}

func (c *InstanceController) RegisterRoutes(router *Router) {
	router.GET("/api/instances", c.List)
	router.POST("/api/instances", c.Create)
	router.GET("/api/instances/:name", c.Get)
	router.PUT("/api/instances/:name", c.Update)
	router.PUT("/api/instances/:name/edit", c.Edit)
	router.DELETE("/api/instances/:name", c.Delete)
	router.POST("/api/instances/:name/start", c.Start)
	router.POST("/api/instances/:name/stop", c.Stop)
	router.POST("/api/instances/:name/reload", c.Reload)
	router.POST("/api/instances/:name/restart", c.Restart)
	router.GET("/api/instances/:name/stats", c.Stats)
	router.GET("/api/instances/:name/logs", c.Logs)
	router.GET("/api/instances/:name/health", c.InstanceHealth)
}
