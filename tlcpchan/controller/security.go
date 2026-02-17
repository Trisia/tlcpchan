package controller

import (
	"encoding/json"
	"net/http"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/security"
	"github.com/Trisia/tlcpchan/security/keystore"
)

// SecurityController 安全参数管理控制器
// 负责管理 keystore 和根证书，同时负责更新配置文件
type SecurityController struct {
	keyStoreMgr *security.KeyStoreManager // keystore 管理器
	rootCertMgr *security.RootCertManager // 根证书管理器
	cfg         *config.Config            // 全局配置
	configPath  string                    // 配置文件路径
}

// NewSecurityController 创建安全参数管理控制器
// 参数：
//   - keyStoreMgr: keystore 管理器
//   - rootCertMgr: 根证书管理器
//   - cfg: 全局配置对象
//   - configPath: 配置文件路径
//
// 返回：
//   - *SecurityController: 新的控制器实例
func NewSecurityController(keyStoreMgr *security.KeyStoreManager, rootCertMgr *security.RootCertManager, cfg *config.Config, configPath string) *SecurityController {
	return &SecurityController{
		keyStoreMgr: keyStoreMgr,
		rootCertMgr: rootCertMgr,
		cfg:         cfg,
		configPath:  configPath,
	}
}

// RegisterRoutes 注册路由
func (c *SecurityController) RegisterRoutes(r *Router) {
	r.GET("/api/security/keystores", c.ListKeyStores)
	r.POST("/api/security/keystores", c.CreateKeyStore)
	r.GET("/api/security/keystores/:name", c.GetKeyStore)
	r.DELETE("/api/security/keystores/:name", c.DeleteKeyStore)
	r.POST("/api/security/keystores/:name/reload", c.ReloadKeyStore)

	r.GET("/api/security/rootcerts", c.ListRootCerts)
	r.POST("/api/security/rootcerts", c.AddRootCert)
	r.GET("/api/security/rootcerts/:filename", c.GetRootCert)
	r.DELETE("/api/security/rootcerts/:filename", c.DeleteRootCert)
	r.POST("/api/security/rootcerts/reload", c.ReloadRootCerts)
}

/**
 * @api {get} /api/security/keystores 列出所有 keystore
 * @apiName ListKeyStores
 * @apiGroup Security-KeyStore
 * @apiVersion 1.0.0
 *
 * @apiDescription 获取系统中所有配置的密钥库（keystore）列表
 *
 * @apiSuccess {Object[]} - keystore 列表数组
 * @apiSuccess {String} -.name keystore 名称，唯一标识符
 * @apiSuccess {String} -.type keystore 类型，可选值："tlcp"（国密）、"tls"（标准）
 * @apiSuccess {String} -.loaderType 加载器类型，可选值："file"（文件）、"named"（命名）、"skf"（SKF设备）、"sdf"（SDF设备）
 * @apiSuccess {Object} -.params 加载器参数，键值对形式，具体内容取决于加载器类型
 * @apiSuccess {Boolean} -.protected 是否受保护，true 表示需要密码访问
 * @apiSuccess {String} -.createdAt 创建时间，ISO 8601 格式
 * @apiSuccess {String} -.updatedAt 更新时间，ISO 8601 格式
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     [
 *       {
 *         "name": "tlcp-server",
 *         "type": "tlcp",
 *         "loaderType": "file",
 *         "params": {
 *           "sign-cert": "sign.crt",
 *           "sign-key": "sign.key",
 *           "enc-cert": "enc.crt",
 *           "enc-key": "enc.key"
 *         },
 *         "protected": false,
 *         "createdAt": "2024-01-01T00:00:00Z",
 *         "updatedAt": "2024-01-01T00:00:00Z"
 *       },
 *       {
 *         "name": "tls-client",
 *         "type": "tls",
 *         "loaderType": "file",
 *         "params": {
 *           "cert": "client.crt",
 *           "key": "client.key"
 *         },
 *         "protected": false,
 *         "createdAt": "2024-01-01T00:00:00Z",
 *         "updatedAt": "2024-01-01T00:00:00Z"
 *       }
 *     ]
 */
func (c *SecurityController) ListKeyStores(w http.ResponseWriter, r *http.Request) {
	keyStores := c.keyStoreMgr.List()
	Success(w, keyStores)
}

/**
 * @api {post} /api/security/keystores 创建 keystore
 * @apiName CreateKeyStore
 * @apiGroup Security-KeyStore
 * @apiVersion 1.0.0
 *
 * @apiDescription 创建新的密钥库（keystore），创建成功后会自动更新配置文件
 *
 * @apiBody {String} name keystore 名称，唯一标识符，只能包含字母、数字、下划线和连字符
 * @apiBody {String} loaderType 加载器类型，可选值："file"（文件加载器）、"named"（命名加载器）、"skf"（SKF设备）、"sdf"（SDF设备）
 * @apiBody {Object} params 加载器参数，键值对形式，具体内容取决于加载器类型：
 *   - file 加载器：
 *     - TLCP: {"sign-cert": "...", "sign-key": "...", "enc-cert": "...", "enc-key": "..."}
 *     - TLS: {"cert": "...", "key": "..."}
 * @apiBody {Boolean} [protected=false] 是否受保护，true 表示需要密码访问
 *
 * @apiSuccess {String} name keystore 名称
 * @apiSuccess {String} type keystore 类型，可选值："tlcp"、"tls"
 * @apiSuccess {String} loaderType 加载器类型
 * @apiSuccess {Object} params 加载器参数
 * @apiSuccess {Boolean} protected 是否受保护
 * @apiSuccess {String} createdAt 创建时间，ISO 8601 格式
 * @apiSuccess {String} updatedAt 更新时间，ISO 8601 格式
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "name": "tlcp-server",
 *       "type": "tlcp",
 *       "loaderType": "file",
 *       "params": {
 *         "sign-cert": "sign.crt",
 *         "sign-key": "sign.key",
 *         "enc-cert": "enc.crt",
 *         "enc-key": "enc.key"
 *       },
 *       "protected": false,
 *       "createdAt": "2024-01-01T00:00:00Z",
 *       "updatedAt": "2024-01-01T00:00:00Z"
 *     }
 *
 * @apiParamExample {json} Request-Example (TLCP):
 *     {
 *       "name": "tlcp-server",
 *       "loaderType": "file",
 *       "params": {
 *         "sign-cert": "sign.crt",
 *         "sign-key": "sign.key",
 *         "enc-cert": "enc.crt",
 *         "enc-key": "enc.key"
 *       },
 *       "protected": false
 *     }
 *
 * @apiParamExample {json} Request-Example (TLS):
 *     {
 *       "name": "tls-client",
 *       "loaderType": "file",
 *       "params": {
 *         "cert": "client.crt",
 *         "key": "client.key"
 *       },
 *       "protected": false
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     无效的请求: json: cannot unmarshal string into Go value
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     名称不能为空
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     创建失败: 具体错误信息
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     保存配置失败: 具体错误信息
 */
func (c *SecurityController) CreateKeyStore(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name       string              `json:"name"`
		LoaderType keystore.LoaderType `json:"loaderType"`
		Params     map[string]string   `json:"params"`
		Protected  bool                `json:"protected"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "无效的请求: "+err.Error())
		return
	}

	if req.Name == "" {
		BadRequest(w, "名称不能为空")
		return
	}

	info, err := c.keyStoreMgr.Create(req.Name, req.LoaderType, req.Params, req.Protected)
	if err != nil {
		InternalError(w, "创建失败: "+err.Error())
		return
	}

	c.cfg.KeyStores = append(c.cfg.KeyStores, config.KeyStoreConfig{
		Name:   req.Name,
		Type:   req.LoaderType,
		Params: req.Params,
	})

	if err := config.Save(c.configPath, c.cfg); err != nil {
		InternalError(w, "保存配置失败: "+err.Error())
		return
	}

	Success(w, info)
}

/**
 * @api {get} /api/security/keystores/:name 获取 keystore 详情
 * @apiName GetKeyStore
 * @apiGroup Security-KeyStore
 * @apiVersion 1.0.0
 *
 * @apiDescription 获取指定密钥库（keystore）的详细信息
 *
 * @apiParam {String} name keystore 名称（路径参数），唯一标识符
 *
 * @apiSuccess {String} name keystore 名称
 * @apiSuccess {String} type keystore 类型，可选值："tlcp"、"tls"
 * @apiSuccess {String} loaderType 加载器类型，可选值："file"、"named"、"skf"、"sdf"
 * @apiSuccess {Object} params 加载器参数
 * @apiSuccess {Boolean} protected 是否受保护
 * @apiSuccess {String} createdAt 创建时间，ISO 8601 格式
 * @apiSuccess {String} updatedAt 更新时间，ISO 8601 格式
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "name": "tlcp-server",
 *       "type": "tlcp",
 *       "loaderType": "file",
 *       "params": {
 *         "sign-cert": "sign.crt",
 *         "sign-key": "sign.key",
 *         "enc-cert": "enc.crt",
 *         "enc-key": "enc.key"
 *       },
 *       "protected": false,
 *       "createdAt": "2024-01-01T00:00:00Z",
 *       "updatedAt": "2024-01-01T00:00:00Z"
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     keystore 不存在
 */
func (c *SecurityController) GetKeyStore(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	info, err := c.keyStoreMgr.Get(name)
	if err != nil {
		NotFound(w, "keystore 不存在")
		return
	}
	Success(w, info)
}

/**
 * @api {delete} /api/security/keystores/:name 删除 keystore
 * @apiName DeleteKeyStore
 * @apiGroup Security-KeyStore
 * @apiVersion 1.0.0
 *
 * @apiDescription 删除指定的密钥库（keystore），删除后会自动更新配置文件
 *
 * @apiParam {String} name keystore 名称（路径参数），唯一标识符
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     null
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     删除失败: 具体错误信息
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     保存配置失败: 具体错误信息
 */
func (c *SecurityController) DeleteKeyStore(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	if err := c.keyStoreMgr.Delete(name); err != nil {
		InternalError(w, "删除失败: "+err.Error())
		return
	}

	newKeyStores := make([]config.KeyStoreConfig, 0, len(c.cfg.KeyStores))
	for _, ks := range c.cfg.KeyStores {
		if ks.Name != name {
			newKeyStores = append(newKeyStores, ks)
		}
	}
	c.cfg.KeyStores = newKeyStores

	if err := config.Save(c.configPath, c.cfg); err != nil {
		InternalError(w, "保存配置失败: "+err.Error())
		return
	}

	Success(w, nil)
}

/**
 * @api {post} /api/security/keystores/:name/reload 重载 keystore
 * @apiName ReloadKeyStore
 * @apiGroup Security-KeyStore
 * @apiVersion 1.0.0
 *
 * @apiDescription 重新加载指定的密钥库（keystore），从数据源重新读取证书和密钥
 *
 * @apiParam {String} name keystore 名称（路径参数），唯一标识符
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     null
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     重载失败: 具体错误信息
 */
func (c *SecurityController) ReloadKeyStore(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	if err := c.keyStoreMgr.Reload(name); err != nil {
		InternalError(w, "重载失败: "+err.Error())
		return
	}
	Success(w, nil)
}

/**
 * @api {get} /api/security/rootcerts 列出所有根证书
 * @apiName ListRootCerts
 * @apiGroup Security-RootCert
 * @apiVersion 1.0.0
 *
 * @apiDescription 获取系统中所有已加载的根证书列表
 *
 * @apiSuccess {Object[]} - 根证书列表数组
 * @apiSuccess {String} -.filename 证书文件名，唯一标识符
 * @apiSuccess {String} -.subject 证书主题（Subject）
 * @apiSuccess {String} -.issuer 证书颁发者（Issuer）
 * @apiSuccess {String} -.notAfter 证书过期时间，ISO 8601 格式
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     [
 *       {
 *         "filename": "root-ca.crt",
 *         "subject": "CN=Root CA, O=Example Org",
 *         "issuer": "CN=Root CA, O=Example Org",
 *         "notAfter": "2030-01-01T00:00:00Z"
 *       },
 *       {
 *         "filename": "sm2-root-ca.crt",
 *         "subject": "CN=SM2 Root CA, O=Example Org",
 *         "issuer": "CN=SM2 Root CA, O=Example Org",
 *         "notAfter": "2030-01-01T00:00:00Z"
 *       }
 *     ]
 */
func (c *SecurityController) ListRootCerts(w http.ResponseWriter, r *http.Request) {
	certs := c.rootCertMgr.List()
	Success(w, certs)
}

/**
 * @api {post} /api/security/rootcerts 添加根证书
 * @apiName AddRootCert
 * @apiGroup Security-RootCert
 * @apiVersion 1.0.0
 *
 * @apiDescription 上传并添加新的根证书到系统中
 *
 * @apiUse MultipartForm
 *
 * @apiParam {String} filename 文件名，表单字段，用于指定保存的证书文件名
 * @apiParam {File} cert 证书文件，表单字段，PEM 或 DER 格式的证书文件，最大 10MB
 *
 * @apiSuccess {String} filename 证书文件名
 * @apiSuccess {String} subject 证书主题（Subject）
 * @apiSuccess {String} issuer 证书颁发者（Issuer）
 * @apiSuccess {String} notAfter 证书过期时间，ISO 8601 格式
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "filename": "root-ca.crt",
 *       "subject": "CN=Root CA, O=Example Org",
 *       "issuer": "CN=Root CA, O=Example Org",
 *       "notAfter": "2030-01-01T00:00:00Z"
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     解析表单失败: 具体错误信息
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     文件名不能为空
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     证书文件不能为空: 具体错误信息
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     添加失败: 具体错误信息
 */
func (c *SecurityController) AddRootCert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		BadRequest(w, "解析表单失败: "+err.Error())
		return
	}

	filename := r.FormValue("filename")
	if filename == "" {
		BadRequest(w, "文件名不能为空")
		return
	}

	file, _, err := r.FormFile("cert")
	if err != nil {
		BadRequest(w, "证书文件不能为空: "+err.Error())
		return
	}
	defer file.Close()

	certData := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			certData = append(certData, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	cert, err := c.rootCertMgr.Add(filename, certData)
	if err != nil {
		InternalError(w, "添加失败: "+err.Error())
		return
	}

	Success(w, cert)
}

/**
 * @api {get} /api/security/rootcerts/:filename 获取根证书详情
 * @apiName GetRootCert
 * @apiGroup Security-RootCert
 * @apiVersion 1.0.0
 *
 * @apiDescription 获取指定根证书的详细信息
 *
 * @apiParam {String} filename 证书文件名（路径参数），唯一标识符
 *
 * @apiSuccess {String} filename 证书文件名
 * @apiSuccess {String} subject 证书主题（Subject）
 * @apiSuccess {String} issuer 证书颁发者（Issuer）
 * @apiSuccess {String} notAfter 证书过期时间，ISO 8601 格式
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "filename": "root-ca.crt",
 *       "subject": "CN=Root CA, O=Example Org",
 *       "issuer": "CN=Root CA, O=Example Org",
 *       "notAfter": "2030-01-01T00:00:00Z"
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     根证书不存在
 */
func (c *SecurityController) GetRootCert(w http.ResponseWriter, r *http.Request) {
	filename := PathParam(r, "filename")
	cert, err := c.rootCertMgr.Get(filename)
	if err != nil {
		NotFound(w, "根证书不存在")
		return
	}
	Success(w, cert)
}

/**
 * @api {delete} /api/security/rootcerts/:filename 删除根证书
 * @apiName DeleteRootCert
 * @apiGroup Security-RootCert
 * @apiVersion 1.0.0
 *
 * @apiDescription 删除指定的根证书
 *
 * @apiParam {String} filename 证书文件名（路径参数），唯一标识符
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     null
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     删除失败: 具体错误信息
 */
func (c *SecurityController) DeleteRootCert(w http.ResponseWriter, r *http.Request) {
	filename := PathParam(r, "filename")
	if err := c.rootCertMgr.Delete(filename); err != nil {
		InternalError(w, "删除失败: "+err.Error())
		return
	}
	Success(w, nil)
}

/**
 * @api {post} /api/security/rootcerts/reload 重载所有根证书
 * @apiName ReloadRootCerts
 * @apiGroup Security-RootCert
 * @apiVersion 1.0.0
 *
 * @apiDescription 重新从证书目录加载所有根证书
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     null
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     重载失败: 具体错误信息
 */
func (c *SecurityController) ReloadRootCerts(w http.ResponseWriter, r *http.Request) {
	if err := c.rootCertMgr.Reload(); err != nil {
		InternalError(w, "重载失败: "+err.Error())
		return
	}
	Success(w, nil)
}
