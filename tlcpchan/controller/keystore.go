package controller

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/security/certgen"
	"github.com/Trisia/tlcpchan/security/keystore"
	"github.com/emmansun/gmsm/smx509"
)

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

// CSRParams 证书请求参数
type CSRParams struct {
	CommonName      string   `json:"commonName"`
	Country         string   `json:"country,omitempty"`
	StateOrProvince string   `json:"stateOrProvince,omitempty"`
	Locality        string   `json:"locality,omitempty"`
	Org             string   `json:"org,omitempty"`
	OrgUnit         string   `json:"orgUnit,omitempty"`
	EmailAddress    string   `json:"emailAddress,omitempty"`
	DNSNames        []string `json:"dnsNames,omitempty"`
	IPAddresses     []string `json:"ipAddresses,omitempty"`
}

// ExportCSRRequest 导出CSR请求
type ExportCSRRequest struct {
	KeyType   keystore.KeyType `json:"keyType"`
	CSRParams CSRParams        `json:"csrParams"`
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

	if err := config.Save(c.cfg); err != nil {
		InternalError(w, "保存配置失败: "+err.Error())
		return
	}

	c.log.Info("创建 keystore: %s", name)
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

	if err := config.Save(c.cfg); err != nil {
		InternalError(w, "保存配置失败: "+err.Error())
		return
	}

	c.log.Info("删除 keystore: %s", name)
	Success(w, nil)
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

	if err := config.Save(c.cfg); err != nil {
		InternalError(w, "保存配置失败: "+err.Error())
		return
	}

	c.log.Info("生成 keystore: %s", req.Name)
	Success(w, info)
}

/**
 * @api {post} /api/security/keystores/:name/export-csr 导出证书请求(CSR)
 * @apiName ExportCSR
 * @apiGroup Security-KeyStore
 * @apiVersion 1.0.0
 *
 * @apiDescription 使用现有密钥生成并导出证书请求文件(CSR)
 *
 * @apiParam {String} name keystore 名称（路径参数），唯一标识符
 *
 * @apiBody {String} keyType 密钥类型，可选值："sign"（签名密钥）、"enc"（加密密钥，仅TLCP有效）
 * @apiBody {Object} csrParams 证书请求参数
 * @apiBody {String} csrParams.commonName 通用名称(CN)
 * @apiBody {String} [csrParams.country] 国家代码(C)，2个字母
 * @apiBody {String} [csrParams.stateOrProvince] 省/州(ST)
 * @apiBody {String} [csrParams.locality] 地区/城市(L)
 * @apiBody {String} [csrParams.org] 组织名称(O)
 * @apiBody {String} [csrParams.orgUnit] 组织单位(OU)
 * @apiBody {String} [csrParams.emailAddress] 邮箱地址
 * @apiBody {String[]} [csrParams.dnsNames] DNS主题备用名称列表
 * @apiBody {String[]} [csrParams.ipAddresses] IP主题备用名称列表
 *
 * @apiSuccess {File} - PEM格式的证书请求文件(.csr)
 *
 * @apiParamExample {json} Request-Example:
 *     {
 *       "keyType": "sign",
 *       "csrParams": {
 *         "commonName": "example.com",
 *         "country": "CN",
 *         "stateOrProvince": "Beijing",
 *         "locality": "Beijing",
 *         "org": "Example Org",
 *         "orgUnit": "IT",
 *         "emailAddress": "admin@example.com",
 *         "dnsNames": ["example.com", "www.example.com"],
 *         "ipAddresses": ["192.168.1.1"]
 *       }
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     keystore 不存在
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     commonName不能为空
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     生成CSR失败: 具体错误信息
 */
func (c *SecurityController) ExportCSR(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	ks, err := c.keyStoreMgr.GetKeyStore(name)
	if err != nil {
		NotFound(w, "keystore 不存在")
		return
	}

	var req ExportCSRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "无效的请求: "+err.Error())
		return
	}

	if req.CSRParams.CommonName == "" {
		BadRequest(w, "commonName不能为空")
		return
	}

	// 根据 KeyStore 类型选择获取私钥的方式
	var privateKey interface{}
	var keyStoreType keystore.KeyStoreType

	ksType := ks.Type()
	keyStoreType = ksType

	if ksType == keystore.KeyStoreTypeTLCP {
		certs, err := ks.TLCPCertificate()
		if err != nil {
			InternalError(w, "获取证书失败: "+err.Error())
			return
		}
		if len(certs) == 0 {
			InternalError(w, "证书不存在")
			return
		}
		if req.KeyType == keystore.KeyTypeEnc && len(certs) > 1 {
			privateKey = certs[1].PrivateKey
		} else {
			privateKey = certs[0].PrivateKey
		}
	} else {
		cert, err := ks.TLSCertificate()
		if err != nil {
			InternalError(w, "获取证书失败: "+err.Error())
			return
		}
		if cert == nil {
			InternalError(w, "证书不存在")
			return
		}
		privateKey = cert.PrivateKey
	}

	// 组装 CSR 模板
	subject := pkix.Name{
		CommonName:         req.CSRParams.CommonName,
		Country:            []string{req.CSRParams.Country},
		Province:           []string{req.CSRParams.StateOrProvince},
		Locality:           []string{req.CSRParams.Locality},
		Organization:       []string{req.CSRParams.Org},
		OrganizationalUnit: []string{req.CSRParams.OrgUnit},
	}

	template := &x509.CertificateRequest{
		Subject:            subject,
		DNSNames:           req.CSRParams.DNSNames,
		IPAddresses:        make([]net.IP, 0, len(req.CSRParams.IPAddresses)),
		SignatureAlgorithm: x509.UnknownSignatureAlgorithm,
	}

	for _, ipStr := range req.CSRParams.IPAddresses {
		if ip := net.ParseIP(ipStr); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		}
	}

	// 根据 keystore 类型选择使用 smx509 还是 x509 生成 CSR
	var derBytes []byte
	if keyStoreType == keystore.KeyStoreTypeTLCP {
		derBytes, err = smx509.CreateCertificateRequest(rand.Reader, template, privateKey)
		if err != nil {
			InternalError(w, "生成SM2证书请求失败: "+err.Error())
			return
		}
	} else {
		derBytes, err = x509.CreateCertificateRequest(rand.Reader, template, privateKey)
		if err != nil {
			InternalError(w, "生成证书请求失败: "+err.Error())
			return
		}
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: derBytes,
	})

	timestamp := time.Now().Format("20060102150405")
	var keyTypeSuffix string
	if req.KeyType == "" {
		keyTypeSuffix = "default"
	} else {
		keyTypeSuffix = string(req.KeyType)
	}
	filename := fmt.Sprintf("%s-%s-%s.csr", name, keyTypeSuffix, timestamp)

	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Write(pemBytes)
}
