package controller

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/security"
	"github.com/Trisia/tlcpchan/security/certgen"
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
	r.POST("/api/security/keystores/generate", c.GenerateKeyStore)
	r.GET("/api/security/keystores/:name", c.GetKeyStore)
	r.DELETE("/api/security/keystores/:name", c.DeleteKeyStore)
	r.POST("/api/security/keystores/:name/reload", c.ReloadKeyStore)

	r.GET("/api/security/rootcerts", c.ListRootCerts)
	r.POST("/api/security/rootcerts", c.AddRootCert)
	r.POST("/api/security/rootcerts/generate", c.GenerateRootCA)
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
	var name string
	var loaderType keystore.LoaderType
	var params map[string]string
	var protected bool

	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
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

		name = req.Name
		loaderType = req.LoaderType
		params = req.Params
		protected = req.Protected
	} else {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			BadRequest(w, "解析表单失败: "+err.Error())
			return
		}

		name = r.FormValue("name")
		loaderType = keystore.LoaderType(r.FormValue("loaderType"))
		protected = r.FormValue("protected") == "true"

		if loaderType == keystore.LoaderTypeFile {
			keystoreDir := filepath.Join(c.cfg.WorkDir, "keystores")
			params = make(map[string]string)

			if signCertFile, _, err := r.FormFile("sign-cert"); err == nil {
				defer signCertFile.Close()
				signCertData := make([]byte, 0)
				buf := make([]byte, 1024)
				for {
					n, err := signCertFile.Read(buf)
					if n > 0 {
						signCertData = append(signCertData, buf[:n]...)
					}
					if err != nil {
						break
					}
				}
				signCertPath := filepath.Join(keystoreDir, name+"-sign.crt")
				if err := os.WriteFile(signCertPath, signCertData, 0644); err != nil {
					InternalError(w, "保存签名证书失败: "+err.Error())
					return
				}
				params["sign-cert"] = "./keystores/" + name + "-sign.crt"
			}

			if signKeyFile, _, err := r.FormFile("sign-key"); err == nil {
				defer signKeyFile.Close()
				signKeyData := make([]byte, 0)
				buf := make([]byte, 1024)
				for {
					n, err := signKeyFile.Read(buf)
					if n > 0 {
						signKeyData = append(signKeyData, buf[:n]...)
					}
					if err != nil {
						break
					}
				}
				signKeyPath := filepath.Join(keystoreDir, name+"-sign.key")
				if err := os.WriteFile(signKeyPath, signKeyData, 0600); err != nil {
					InternalError(w, "保存签名密钥失败: "+err.Error())
					return
				}
				params["sign-key"] = "./keystores/" + name + "-sign.key"
			}

			if encCertFile, _, err := r.FormFile("enc-cert"); err == nil {
				defer encCertFile.Close()
				encCertData := make([]byte, 0)
				buf := make([]byte, 1024)
				for {
					n, err := encCertFile.Read(buf)
					if n > 0 {
						encCertData = append(encCertData, buf[:n]...)
					}
					if err != nil {
						break
					}
				}
				encCertPath := filepath.Join(keystoreDir, name+"-enc.crt")
				if err := os.WriteFile(encCertPath, encCertData, 0644); err != nil {
					InternalError(w, "保存加密证书失败: "+err.Error())
					return
				}
				params["enc-cert"] = "./keystores/" + name + "-enc.crt"
			}

			if encKeyFile, _, err := r.FormFile("enc-key"); err == nil {
				defer encKeyFile.Close()
				encKeyData := make([]byte, 0)
				buf := make([]byte, 1024)
				for {
					n, err := encKeyFile.Read(buf)
					if n > 0 {
						encKeyData = append(encKeyData, buf[:n]...)
					}
					if err != nil {
						break
					}
				}
				encKeyPath := filepath.Join(keystoreDir, name+"-enc.key")
				if err := os.WriteFile(encKeyPath, encKeyData, 0600); err != nil {
					InternalError(w, "保存加密密钥失败: "+err.Error())
					return
				}
				params["enc-key"] = "./keystores/" + name + "-enc.key"
			}
		} else {
			paramsStr := r.FormValue("params")
			if paramsStr != "" {
				if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
					BadRequest(w, "无效的 params: "+err.Error())
					return
				}
			}
		}
	}

	if name == "" {
		BadRequest(w, "名称不能为空")
		return
	}

	info, err := c.keyStoreMgr.Create(name, loaderType, params, protected)
	if err != nil {
		InternalError(w, "创建失败: "+err.Error())
		return
	}

	c.cfg.KeyStores = append(c.cfg.KeyStores, config.KeyStoreConfig{
		Name:   name,
		Type:   loaderType,
		Params: params,
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
 * @apiBody {String} filename 文件名，表单字段，用于指定保存的证书文件名
 * @apiBody {File} cert 证书文件，表单字段，PEM 或 DER 格式的证书文件，最大 10MB
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
 * @apiDescription 重新加载所有根证书，从信任证书目录重新读取
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

// GenerateKeyStoreRequest 生成 keystore 请求
type GenerateKeyStoreRequest struct {
	Name           string                     `json:"name"`
	Type           keystore.KeyStoreType      `json:"type"`
	Protected      bool                       `json:"protected"`
	CertConfig     GenerateKeyStoreCertConfig `json:"certConfig"`
	SignerKeyStore string                     `json:"signerKeyStore,omitempty"`
}

// GenerateKeyStoreCertConfig 证书生成配置
type GenerateKeyStoreCertConfig struct {
	CommonName      string   `json:"commonName"`
	Country         string   `json:"country,omitempty"`
	StateOrProvince string   `json:"stateOrProvince,omitempty"`
	Locality        string   `json:"locality,omitempty"`
	Org             string   `json:"org,omitempty"`
	OrgUnit         string   `json:"orgUnit,omitempty"`
	EmailAddress    string   `json:"emailAddress,omitempty"`
	Years           int      `json:"years,omitempty"`
	Days            int      `json:"days,omitempty"`
	KeyAlgorithm    string   `json:"keyAlgorithm,omitempty"`
	KeyBits         int      `json:"keyBits,omitempty"`
	DNSNames        []string `json:"dnsNames,omitempty"`
	IPAddresses     []string `json:"ipAddresses,omitempty"`
}

/**
 * @api {post} /api/security/keystores/generate 生成 keystore（含证书）
 * @apiName GenerateKeyStore
 * @apiGroup Security-KeyStore
 * @apiVersion 1.0.0
 *
 * @apiDescription 生成新的密钥库（keystore），包含自动生成的证书和密钥，支持 TLCP 和 TLS 两种类型
 *
 * @apiBody {String} name keystore 名称，唯一标识符
 * @apiBody {String} type keystore 类型，可选值："tlcp"（国密）、"tls"（标准）
 * @apiBody {Boolean} [protected=false] 是否受保护，true 表示需要密码访问
 * @apiBody {Object} certConfig 证书生成配置
 * @apiBody {String} certConfig.commonName 证书通用名称（CN）
 * @apiBody {String} certConfig.org 组织名称（O）
 * @apiBody {String} certConfig.orgUnit 组织单位（OU）
 * @apiBody {Number} certConfig.years 证书有效期（年）
 * @apiBody {String} [signerKeyStore] 用于签发的 keystore 名称（暂未实现）
 *
 * @apiSuccess {String} name keystore 名称
 * @apiSuccess {String} type keystore 类型
 * @apiSuccess {String} loaderType 加载器类型，固定为 "file"
 * @apiSuccess {Object} params 加载器参数
 * @apiSuccess {Boolean} protected 是否受保护
 * @apiSuccess {String} createdAt 创建时间，ISO 8601 格式
 * @apiSuccess {String} updatedAt 更新时间，ISO 8601 格式
 *
 * @apiSuccessExample {json} Success-Response (TLCP):
 *     HTTP/1.1 200 OK
 *     {
 *       "name": "tlcp-server",
 *       "type": "tlcp",
 *       "loaderType": "file",
 *       "params": {
 *         "sign-cert": "./keystores/tlcp-server-sign.crt",
 *         "sign-key": "./keystores/tlcp-server-sign.key",
 *         "enc-cert": "./keystores/tlcp-server-enc.crt",
 *         "enc-key": "./keystores/tlcp-server-enc.key"
 *       },
 *       "protected": false,
 *       "createdAt": "2024-01-01T00:00:00Z",
 *       "updatedAt": "2024-01-01T00:00:00Z"
 *     }
 *
 * @apiSuccessExample {json} Success-Response (TLS):
 *     HTTP/1.1 200 OK
 *     {
 *       "name": "tls-server",
 *       "type": "tls",
 *       "loaderType": "file",
 *       "params": {
 *         "sign-cert": "./keystores/tls-server.crt",
 *         "sign-key": "./keystores/tls-server.key"
 *       },
 *       "protected": false,
 *       "createdAt": "2024-01-01T00:00:00Z",
 *       "updatedAt": "2024-01-01T00:00:00Z"
 *     }
 *
 * @apiParamExample {json} Request-Example (TLCP):
 *     {
 *       "name": "tlcp-server",
 *       "type": "tlcp",
 *       "protected": false,
 *       "certConfig": {
 *         "commonName": "example.com",
 *         "org": "Example Org",
 *         "orgUnit": "IT",
 *         "years": 1
 *       }
 *     }
 *
 * @apiParamExample {json} Request-Example (TLS):
 *     {
 *       "name": "tls-server",
 *       "type": "tls",
 *       "protected": false,
 *       "certConfig": {
 *         "commonName": "example.com",
 *         "org": "Example Org",
 *         "orgUnit": "IT",
 *         "years": 1
 *       }
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     无效的请求: json: cannot unmarshal string into Go value
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     名称不能为空
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     类型不能为空
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     不支持的 keystore 类型: xxx
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     生成根证书失败: 具体错误信息
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     保存配置失败: 具体错误信息
 */
func (c *SecurityController) GenerateKeyStore(w http.ResponseWriter, r *http.Request) {
	var req GenerateKeyStoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "无效的请求: "+err.Error())
		return
	}

	if req.Name == "" {
		BadRequest(w, "名称不能为空")
		return
	}
	if req.Type == "" {
		BadRequest(w, "类型不能为空")
		return
	}

	keystoreDir := filepath.Join(c.cfg.WorkDir, "keystores")

	var params map[string]string

	if req.Type == keystore.KeyStoreTypeTLCP {
		signCertPath := filepath.Join(keystoreDir, req.Name+"-sign.crt")
		signKeyPath := filepath.Join(keystoreDir, req.Name+"-sign.key")
		encCertPath := filepath.Join(keystoreDir, req.Name+"-enc.crt")
		encKeyPath := filepath.Join(keystoreDir, req.Name+"-enc.key")

		var err error
		var signerCertPath, signerKeyPath string

		if req.SignerKeyStore != "" {
			BadRequest(w, "使用签发者证书功能暂未实现")
			return
		} else {
			caCert, err := certgen.GenerateTLCPRootCA(certgen.CertGenConfig{
				Type:            certgen.CertTypeRootCA,
				CommonName:      req.Name + "-ca",
				Country:         req.CertConfig.Country,
				StateOrProvince: req.CertConfig.StateOrProvince,
				Locality:        req.CertConfig.Locality,
				Org:             req.CertConfig.Org,
				OrgUnit:         req.CertConfig.OrgUnit,
				EmailAddress:    req.CertConfig.EmailAddress,
				Years:           10,
				Days:            0,
			})
			if err != nil {
				InternalError(w, "生成根证书失败: "+err.Error())
				return
			}
			signerCertPath = filepath.Join(keystoreDir, req.Name+"-ca.crt")
			signerKeyPath = filepath.Join(keystoreDir, req.Name+"-ca.key")
			if err := certgen.SaveCertToFile(caCert.CertPEM, caCert.KeyPEM, signerCertPath, signerKeyPath); err != nil {
				InternalError(w, "保存根证书失败: "+err.Error())
				return
			}
		}

		signCfg := certgen.CertGenConfig{
			Type:            certgen.CertTypeTLCPSign,
			CommonName:      req.CertConfig.CommonName + "-sign",
			Country:         req.CertConfig.Country,
			StateOrProvince: req.CertConfig.StateOrProvince,
			Locality:        req.CertConfig.Locality,
			Org:             req.CertConfig.Org,
			OrgUnit:         req.CertConfig.OrgUnit,
			EmailAddress:    req.CertConfig.EmailAddress,
			Years:           req.CertConfig.Years,
			Days:            req.CertConfig.Days,
			DNSNames:        req.CertConfig.DNSNames,
			IPAddresses:     req.CertConfig.IPAddresses,
		}
		encCfg := certgen.CertGenConfig{
			Type:            certgen.CertTypeTLCPEnc,
			CommonName:      req.CertConfig.CommonName + "-enc",
			Country:         req.CertConfig.Country,
			StateOrProvince: req.CertConfig.StateOrProvince,
			Locality:        req.CertConfig.Locality,
			Org:             req.CertConfig.Org,
			OrgUnit:         req.CertConfig.OrgUnit,
			EmailAddress:    req.CertConfig.EmailAddress,
			Years:           req.CertConfig.Years,
			Days:            req.CertConfig.Days,
			DNSNames:        req.CertConfig.DNSNames,
			IPAddresses:     req.CertConfig.IPAddresses,
		}

		signerX509Cert, signerPrivKey, err := certgen.LoadTLCPCertFromFile(
			signerCertPath,
			signerKeyPath,
		)
		if err != nil {
			InternalError(w, "加载签发者证书失败: "+err.Error())
			return
		}

		signCert, encCert, err := certgen.GenerateTLCPPair(
			signerX509Cert, signerPrivKey, signCfg, encCfg,
		)
		if err != nil {
			InternalError(w, "生成TLCP证书失败: "+err.Error())
			return
		}

		if err := certgen.SaveCertToFile(signCert.CertPEM, signCert.KeyPEM, signCertPath, signKeyPath); err != nil {
			InternalError(w, "保存签名证书失败: "+err.Error())
			return
		}
		if err := certgen.SaveCertToFile(encCert.CertPEM, encCert.KeyPEM, encCertPath, encKeyPath); err != nil {
			InternalError(w, "保存加密证书失败: "+err.Error())
			return
		}

		params = map[string]string{
			"sign-cert": "./keystores/" + req.Name + "-sign.crt",
			"sign-key":  "./keystores/" + req.Name + "-sign.key",
			"enc-cert":  "./keystores/" + req.Name + "-enc.crt",
			"enc-key":   "./keystores/" + req.Name + "-enc.key",
		}
	} else if req.Type == keystore.KeyStoreTypeTLS {
		certPath := filepath.Join(keystoreDir, req.Name+".crt")
		keyPath := filepath.Join(keystoreDir, req.Name+".key")

		var keyAlg certgen.KeyAlgorithm
		switch req.CertConfig.KeyAlgorithm {
		case "rsa":
			keyAlg = certgen.KeyAlgorithmRSA
		case "ecdsa":
			keyAlg = certgen.KeyAlgorithmECDSA
		default:
			keyAlg = certgen.KeyAlgorithmECDSA
		}

		tlsCfg := certgen.CertGenConfig{
			Type:            certgen.CertTypeTLS,
			CommonName:      req.CertConfig.CommonName,
			Country:         req.CertConfig.Country,
			StateOrProvince: req.CertConfig.StateOrProvince,
			Locality:        req.CertConfig.Locality,
			Org:             req.CertConfig.Org,
			OrgUnit:         req.CertConfig.OrgUnit,
			EmailAddress:    req.CertConfig.EmailAddress,
			Years:           req.CertConfig.Years,
			Days:            req.CertConfig.Days,
			KeyAlgorithm:    keyAlg,
			KeyBits:         req.CertConfig.KeyBits,
			DNSNames:        req.CertConfig.DNSNames,
			IPAddresses:     req.CertConfig.IPAddresses,
		}

		signerCert, err := certgen.GenerateTLSRootCA(certgen.CertGenConfig{
			Type:            certgen.CertTypeRootCA,
			CommonName:      req.Name + "-ca",
			Country:         req.CertConfig.Country,
			StateOrProvince: req.CertConfig.StateOrProvince,
			Locality:        req.CertConfig.Locality,
			Org:             req.CertConfig.Org,
			OrgUnit:         req.CertConfig.OrgUnit,
			EmailAddress:    req.CertConfig.EmailAddress,
			Years:           10,
			Days:            0,
		})
		if err != nil {
			InternalError(w, "生成根证书失败: "+err.Error())
			return
		}

		signerCertPath := filepath.Join(keystoreDir, req.Name+"-ca.crt")
		signerKeyPath := filepath.Join(keystoreDir, req.Name+"-ca.key")
		if err := certgen.SaveCertToFile(signerCert.CertPEM, signerCert.KeyPEM, signerCertPath, signerKeyPath); err != nil {
			InternalError(w, "保存根证书失败: "+err.Error())
			return
		}

		signerX509Cert, signerPrivKey, err := certgen.LoadTLSCertFromFile(signerCertPath, signerKeyPath)
		if err != nil {
			InternalError(w, "加载签发者证书失败: "+err.Error())
			return
		}

		tlsCert, err := certgen.GenerateTLSCert(signerX509Cert, signerPrivKey, tlsCfg)
		if err != nil {
			InternalError(w, "生成TLS证书失败: "+err.Error())
			return
		}

		if err := certgen.SaveCertToFile(tlsCert.CertPEM, tlsCert.KeyPEM, certPath, keyPath); err != nil {
			InternalError(w, "保存证书失败: "+err.Error())
			return
		}

		params = map[string]string{
			"sign-cert": "./keystores/" + req.Name + ".crt",
			"sign-key":  "./keystores/" + req.Name + ".key",
		}
	} else {
		BadRequest(w, "不支持的 keystore 类型: "+string(req.Type))
		return
	}

	info, err := c.keyStoreMgr.Create(req.Name, keystore.LoaderTypeFile, params, req.Protected)
	if err != nil {
		InternalError(w, "创建 keystore 失败: "+err.Error())
		return
	}

	c.cfg.KeyStores = append(c.cfg.KeyStores, config.KeyStoreConfig{
		Name:   req.Name,
		Type:   keystore.LoaderTypeFile,
		Params: params,
	})

	if err := config.Save(c.configPath, c.cfg); err != nil {
		InternalError(w, "保存配置失败: "+err.Error())
		return
	}

	Success(w, info)
}

// GenerateRootCARequest 生成根 CA 请求
type GenerateRootCARequest struct {
	Type            string `json:"type,omitempty"` // 类型：tlcp 或 tls，默认 tlcp
	CommonName      string `json:"commonName"`
	Country         string `json:"country,omitempty"`
	StateOrProvince string `json:"stateOrProvince,omitempty"`
	Locality        string `json:"locality,omitempty"`
	Org             string `json:"org,omitempty"`
	OrgUnit         string `json:"orgUnit,omitempty"`
	EmailAddress    string `json:"emailAddress,omitempty"`
	Years           int    `json:"years,omitempty"`
	Days            int    `json:"days,omitempty"`
}

/**
 * @api {post} /api/security/rootcerts/generate 生成根 CA 证书
 * @apiName GenerateRootCA
 * @apiGroup Security-RootCert
 * @apiVersion 1.0.0
 *
 * @apiDescription 生成新的自签名根 CA 证书，并添加到信任证书列表中，同时创建对应的 keystore
 *
 * @apiBody {String} [commonName="tlcpchan-root-ca"] 证书通用名称（CN）
 * @apiBody {String} [org="tlcpchan"] 组织名称（O）
 * @apiBody {String} [orgUnit] 组织单位（OU）
 * @apiBody {Number} [years=10] 证书有效期（年）
 *
 * @apiSuccess {String} filename 证书文件名
 * @apiSuccess {String} subject 证书主题（Subject）
 * @apiSuccess {String} issuer 证书颁发者（Issuer）
 * @apiSuccess {String} notAfter 证书过期时间，ISO 8601 格式
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "filename": "my-root-ca.crt",
 *       "subject": "CN=my-root-ca, O=My Org",
 *       "issuer": "CN=my-root-ca, O=My Org",
 *       "notAfter": "2034-01-01T00:00:00Z"
 *     }
 *
 * @apiParamExample {json} Request-Example:
 *     {
 *       "commonName": "my-root-ca",
 *       "org": "My Org",
 *       "orgUnit": "IT",
 *       "years": 10
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     无效的请求: json: cannot unmarshal string into Go value
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     生成根证书失败: 具体错误信息
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     添加根证书失败: 具体错误信息
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     保存配置失败: 具体错误信息
 */
func (c *SecurityController) GenerateRootCA(w http.ResponseWriter, r *http.Request) {
	var req GenerateRootCARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "无效的请求: "+err.Error())
		return
	}

	if req.CommonName == "" {
		req.CommonName = "tlcpchan-root-ca"
	}
	if req.Org == "" {
		req.Org = "tlcpchan"
	}
	if req.Years <= 0 {
		req.Years = 10
	}

	var rootCA *certgen.GeneratedCert
	var err error
	cfg := certgen.CertGenConfig{
		Type:            certgen.CertTypeRootCA,
		CommonName:      req.CommonName,
		Country:         req.Country,
		StateOrProvince: req.StateOrProvince,
		Locality:        req.Locality,
		Org:             req.Org,
		OrgUnit:         req.OrgUnit,
		EmailAddress:    req.EmailAddress,
		Years:           req.Years,
		Days:            req.Days,
	}

	if req.Type == "tls" {
		rootCA, err = certgen.GenerateTLSRootCA(cfg)
	} else {
		rootCA, err = certgen.GenerateTLCPRootCA(cfg)
	}

	if err != nil {
		InternalError(w, "生成根证书失败: "+err.Error())
		return
	}

	filename := req.CommonName + ".crt"
	cert, err := c.rootCertMgr.Add(filename, rootCA.CertPEM)
	if err != nil {
		InternalError(w, "添加根证书失败: "+err.Error())
		return
	}

	keystoreDir := filepath.Join(c.cfg.WorkDir, "keystores")
	certPath := filepath.Join(keystoreDir, req.CommonName+".crt")
	keyPath := filepath.Join(keystoreDir, req.CommonName+".key")
	if err := certgen.SaveCertToFile(rootCA.CertPEM, rootCA.KeyPEM, certPath, keyPath); err != nil {
		InternalError(w, "保存证书和密钥失败: "+err.Error())
		return
	}

	params := map[string]string{
		"sign-cert": "./keystores/" + req.CommonName + ".crt",
		"sign-key":  "./keystores/" + req.CommonName + ".key",
	}
	if _, err := c.keyStoreMgr.Create(req.CommonName, keystore.LoaderTypeFile, params, true); err != nil {
		InternalError(w, "创建 keystore 失败: "+err.Error())
		return
	}

	c.cfg.KeyStores = append(c.cfg.KeyStores, config.KeyStoreConfig{
		Name:   req.CommonName,
		Type:   keystore.LoaderTypeFile,
		Params: params,
	})

	if err := config.Save(c.configPath, c.cfg); err != nil {
		InternalError(w, "保存配置失败: "+err.Error())
		return
	}

	Success(w, cert)
}
