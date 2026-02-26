package controller

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	// 如果是 file 类型，验证文件是否存在
	if loaderType == keystore.LoaderTypeFile {
		if err := validateFileParams(c.cfg.WorkDir, params); err != nil {
			BadRequest(w, err.Error())
			return
		}
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
		err = certgen.SaveCertToFile(signerCert.CertPEM, signerCert.KeyPEM, signerCertPath, signerKeyPath)
		if err != nil {
			InternalError(w, "保存根证书失败文件: "+err.Error())
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
	keyStoreType := ks.Type()

	if keyStoreType == keystore.KeyStoreTypeTLCP {
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

/**
 * @api {get} /api/security/keystores/:name/instances 查询引用指定 keystore 的实例列表
 * @apiName GetKeyStoreInstances
 * @apiGroup Security-KeyStore
 * @apiVersion 1.0.0
 *
 * @apiDescription 查询引用指定 keystore 的所有实例列表，包括实例名称、状态和协议信息
 *
 * @apiParam {String} name keystore 名称（路径参数），唯一标识符
 *
 * @apiSuccess {Object[]} - 实例列表数组
 * @apiSuccess {String} -.name 实例名称
 * @apiSuccess {String} -.status 实例状态，可选值：created（已创建）、running（运行中）、stopped（已停止）、error（错误）
 * @apiSuccess {String} -.protocol协议类型，可选值：auto（自动）、tlcp（仅TLCP）、tls（仅TLS）
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     [
 *       {
 *         "name": "tlcp-server",
 *         "status": "running",
 *         "protocol": "tlcp"
 *       },
 *       {
 *         "name": "tls-client",
 *         "status": "stopped",
 *         "protocol": "tls"
 *       }
 *     ]
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     keystore 不存在
 */
func (c *SecurityController) GetKeyStoreInstances(w http.ResponseWriter, r *http.Request) {
	keystoreName := PathParam(r, "name")

	// 先检查 keystore 是否存在
	if _, err := c.keyStoreMgr.Get(keystoreName); err != nil {
		NotFound(w, "keystore 不存在")
		return
	}

	// 遍历所有实例配置，检查是否引用了指定的 keystore
	var result []map[string]interface{}

	// 需要获取运行中实例的状态，这里简化处理，返回配置中的协议信息
	for _, instCfg := range c.cfg.Instances {
		// 检查 tlcp.keystore 是否引用
		if instCfg.TLCP.Keystore != nil {
			if instCfg.TLCP.Keystore.Type == "named" && instCfg.TLCP.Keystore.Name == keystoreName {
				result = append(result, map[string]interface{}{
					"name":     instCfg.Name,
					"status":   getInstanceStateStatus(instCfg.Name, c.cfg.Instances),
					"protocol": instCfg.Protocol,
				})
				continue
			}
			// 检查 file 类型中，params 中是否包含该 keystore 名称的路径
			if instCfg.TLCP.Keystore.Type == "file" {
				for _, param := range instCfg.TLCP.Keystore.Params {
					if containsKeystoreName(param, keystoreName) {
						result = append(result, map[string]interface{}{
							"name":     instCfg.Name,
							"status":   getInstanceStateStatus(instCfg.Name, c.cfg.Instances),
							"protocol": instCfg.Protocol,
						})
						break
					}
				}
			}
		}

		// 检查 tls.keystore 是否引用
		if instCfg.TLS.Keystore != nil {
			if instCfg.TLS.Keystore.Type == "named" && instCfg.TLS.Keystore.Name == keystoreName {
				result = append(result, map[string]interface{}{
					"name":     instCfg.Name,
					"status":   getInstanceStateStatus(instCfg.Name, c.cfg.Instances),
					"protocol": instCfg.Protocol,
				})
				continue
			}
			// 检查 file 类型中，params 中是否包含该 keystore 名称的路径
			if instCfg.TLS.Keystore.Type == "file" {
				for _, param := range instCfg.TLS.Keystore.Params {
					if containsKeystoreName(param, keystoreName) {
						result = append(result, map[string]interface{}{
							"name":     instCfg.Name,
							"status":   getInstanceStateStatus(instCfg.Name, c.cfg.Instances),
							"protocol": instCfg.Protocol,
						})
						break
					}
				}
			}
		}
	}

	Success(w, result)
}

/**
 * containsKeystoreName 检查参数值是否包含 keystore 名称
 *
 * 参数:
 *   - param: 参数值（通常是文件路径）
 *   - keystoreName: keystore 名称
 *
 * 返回:
 *   - bool: 如果参数值包含 keystore 名称则返回 true，否则返回 false
 */
func containsKeystoreName(param, keystoreName string) bool {
	return strings.Contains(param, keystoreName)
}

/**
 * validateFileParams 验证 file 类型 keystore 的所有文件路径是否存在
 *
 * 参数:
 *   - workDir: 工作目录，用于解析相对路径
 *   - params: keystore 参数，包含文件路径
 *
 * 返回:
 *   - error: 如果文件不存在则返回错误信息，否则返回 nil
 */
func validateFileParams(workDir string, params map[string]string) error {
	for _, filePath := range params {
		if filePath == "" {
			continue
		}

		// 解析文件路径，处理相对路径
		var fullPath string
		if strings.HasPrefix(filePath, "./") || strings.HasPrefix(filePath, "/") {
			fullPath = filePath
		} else {
			fullPath = "./" + filePath
		}

		// 如果是相对路径，相对于工作目录解析
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(workDir, fullPath)
		}

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fmt.Errorf("文件 %s 不存在", filePath)
		}
	}
	return nil
}

/**
 * getInstanceStateStatus 获取实例状态
 * 注意：这是一个简化实现，实际状态应该从实例管理器获取
 *
 * 参数:
 *   - instanceName: 实例名称
 *   - instances: 所有实例配置列表
 *
 * 返回:
 *   - string: 实例状态（简化为 "stopped"）
 */
func getInstanceStateStatus(instanceName string, instances []config.InstanceConfig) string {
	// TODO: 从实例管理器获取真实的实例状态
	// 当前返回 "stopped" 作为默认值
	return "stopped"
}

/**
 * @api {put} /api/security/keystores/:name 更新 keystore 参数
 * @apiName UpdateKeystoreParams
 * @apiGroup Security-KeyStore
 * @apiVersion 1.0.0
 *
 * @apiDescription 更新指定 keystore 的参数（如证书和密钥路径的文件路径）
 *
 * @apiParam {String} name keystore 名称（路径参数），唯一标识符
 * @apiBody {Object} params 要更新的参数键值对，只更新提供的字段
 * @apiBody {String} params.sign-cert 签名证书路径（可选）
 * @apiBody {String} params.sign-key 签名密钥路径（可选）
 * @apiBody {String} params.enc-cert 加密证书路径（可选，仅TLCP）
 * @apiBody {String} params.enc-key 加密密钥路径（可选，仅TLCP）
 * @apiBody {String} params.cert 证书路径（可选，仅TLS）
 * @apiBody {String} params.key 密钥路径（可选，仅TLS）
 *
 * @apiSuccess {String} name keystore 名称
 * @apiSuccess {String} type keystore 类型
 * @apiSuccess {String} loaderType 加载器类型
 * @apiSuccess {Object} params 更新后的参数
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
 *         "sign-cert": "./keystores/new-sign.crt",
 *         "sign-key": "./keystores/new-sign.key",
 *         "enc-cert": "./keystores/enc.crt",
 *         "enc-key": "./keystores/enc.key"
 *       },
 *       "protected": false,
 *       "createdAt": "2024-01-01T00:00:00Z",
 *       "updatedAt": "2024-02-25T10:30:00Z"
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     keystore 不存在
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     参数无效: 具体错误信息
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     保存配置失败: 具体错误信息
 */
func (c *SecurityController) UpdateKeystoreParams(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")

	// 检查 keystore 是否存在
	info, err := c.keyStoreMgr.Get(name)
	if err != nil {
		NotFound(w, "keystore 不存在")
		return
	}

	// 检查是否为 protected 状态
	if info.Protected {
		BadRequest(w, "受保护的 keystore 不允许修改")
		return
	}

	// 解析请求体
	var reqBody struct {
		Params map[string]string `json:"params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		BadRequest(w, "无效的请求体: "+err.Error())
		return
	}

	if len(reqBody.Params) == 0 {
		BadRequest(w, "参数不能为空")
		return
	}

	// 验证参数有效性（基本验证）
	for key, value := range reqBody.Params {
		if value == "" {
			BadRequest(w, fmt.Sprintf("参数 %s 不能为空", key))
			return
		}
	}

	// 如果是 file 类型，验证文件是否存在
	if info.LoaderType == keystore.LoaderTypeFile {
		if err := validateFileParams(c.cfg.WorkDir, reqBody.Params); err != nil {
			BadRequest(w, err.Error())
			return
		}
	}

	// 更新 keystore 配置
	// 需要在配置文件中找到对应的 keystore 并更新其参数
	found := false
	for i := range c.cfg.KeyStores {
		if c.cfg.KeyStores[i].Name == name {
			if c.cfg.KeyStores[i].Params == nil {
				c.cfg.KeyStores[i].Params = make(map[string]string)
			}
			// 更新参数
			for key, value := range reqBody.Params {
				c.cfg.KeyStores[i].Params[key] = value
			}
			found = true
			break
		}
	}

	if !found {
		NotFound(w, "配置中未找到 keystore")
		return
	}

	// 保存配置文件
	if err := config.Save(c.cfg); err != nil {
		InternalError(w, "保存配置失败: "+err.Error())
		return
	}

	// 重新加载 keystore
	updatedInfo, err := c.keyStoreMgr.Get(name)
	if err != nil {
		InternalError(w, "重新加载 keystore 失败: "+err.Error())
		return
	}

	c.log.Info("更新 keystore 参数: %s", name)
	Success(w, updatedInfo)
}

/**
 * @api {post} /api/security/keystores/:name/upload 上传更新证书和密钥
 * @apiName UpdateCertificates
 * @apiGroup Security-KeyStore
 * @apiVersion 1.0.0
 *
 * @apiDescription 上传文件以更新指定 keystore 的证书和密钥。对于 TLS 类型，使用 signCert 和 signKey；对于 TLCP 类型，使用 signCert、signKey、encCert 和 encKey
 *
 * @apiParam {String} name keystore 名称（路径参数），唯一标识符
 * @apiBody {File} signCert 签名证书文件（TLS类型时为证书，TLCP类型时为签名证书）
 * @apiBody {File} signKey 签名密钥文件（TLS类型时为密钥，TLCP类型时为签名密钥）
 * @apiBody {File} encCert 加密证书文件（仅TLCP类型有效）
 * @apiBody {File} encKey 加密密钥文件（仅TLCP类型有效）
 *
 * @apiSuccess {String} name keystore 名称
 * @apiSuccess {String} type keystore 类型
 * @apiSuccess {String} loaderType 加载器类型
 * @apiSuccess {Object} params 更新后的参数
 * @apiSuccess {Boolean} protected 是否受保护
 * @apiSuccess {String} createdAt 创建时间，ISO 8601 格式
 * @apiSuccess {String} updatedAt 更新时间，ISO 8601 格式
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     keystore 不存在
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     证书与密钥不匹配
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 403 Forbidden
 *     keystore 受保护，不允许修改
 */
func (c *SecurityController) UpdateCertificates(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")

	// 检查 keystore 是否存在
	info, err := c.keyStoreMgr.Get(name)
	if err != nil {
		NotFound(w, "keystore 不存在")
		return
	}

	// 检查是否为 protected 状态
	if info.Protected {
		BadRequest(w, "受保护的 keystore 不允许修改")
		return
	}

	// 仅支持 file 类型的 keystore
	if info.LoaderType != keystore.LoaderTypeFile {
		BadRequest(w, "只有文件类型的 keystore 支持更新证书和密钥")
		return
	}

	// 解析 multipart 表单
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		BadRequest(w, "解析表单失败: "+err.Error())
		return
	}

	keystoreDir := filepath.Join(c.cfg.WorkDir, "keystores")
	tempDir := filepath.Join(keystoreDir, ".temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		InternalError(w, "创建临时目录失败: "+err.Error())
		return
	}
	defer func() {
		os.RemoveAll(tempDir)
	}()

	// 临时文件路径映射
	tempFiles := make(map[string]string)
	finalFiles := make(map[string]string)

	// 处理文件上传
	isTLCP := info.Type == keystore.KeyStoreTypeTLCP

	if isTLCP {
		// 处理签名证书和密钥
		if signCertFile, signCertData, err := handleFormFile(r, "signCert", tempDir, name, "sign", "crt"); err != nil {
			BadRequest(w, err.Error())
			return
		} else if signCertFile != "" {
			tempFiles["sign-cert"] = signCertFile
			finalFiles["sign-cert"] = filepath.Join(keystoreDir, name+"-sign.crt")

			// 如果同时上传了签名密钥，验证配对
			if signKeyFile, signKeyData, err := handleFormFile(r, "signKey", tempDir, name, "sign", "key"); err != nil {
				BadRequest(w, err.Error())
				return
			} else if signKeyFile != "" {
				tempFiles["sign-key"] = signKeyFile
				finalFiles["sign-key"] = filepath.Join(keystoreDir, name+"-sign.key")

				if err := keystore.VerifyCertificateKeyPair(signCertData, signKeyData, true); err != nil {
					BadRequest(w, "签名证书与密钥不匹配: "+err.Error())
					return
				}
			}
		} else if signKeyFile, _, err := handleFormFile(r, "signKey", tempDir, name, "sign", "key"); err != nil {
			BadRequest(w, err.Error())
			return
		} else if signKeyFile != "" {
			BadRequest(w, "上传签名密钥时必须同时上传签名证书")
			return
		}

		// 处理加密证书和密钥
		if encCertFile, encCertData, err := handleFormFile(r, "encCert", tempDir, name, "enc", "crt"); err != nil {
			BadRequest(w, err.Error())
			return
		} else if encCertFile != "" {
			tempFiles["enc-cert"] = encCertFile
			finalFiles["enc-cert"] = filepath.Join(keystoreDir, name+"-enc.crt")

			// 如果同时上传了加密密钥，验证配对
			if encKeyFile, encKeyData, err := handleFormFile(r, "encKey", tempDir, name, "enc", "key"); err != nil {
				BadRequest(w, err.Error())
				return
			} else if encKeyFile != "" {
				tempFiles["enc-key"] = encKeyFile
				finalFiles["enc-key"] = filepath.Join(keystoreDir, name+"-enc.key")

				if err := keystore.VerifyCertificateKeyPair(encCertData, encKeyData, true); err != nil {
					BadRequest(w, "加密证书与密钥不匹配: "+err.Error())
					return
				}
			}
		} else if encKeyFile, _, err := handleFormFile(r, "encKey", tempDir, name, "enc", "key"); err != nil {
			BadRequest(w, err.Error())
			return
		} else if encKeyFile != "" {
			BadRequest(w, "上传加密密钥时必须同时上传加密证书")
			return
		}
	} else {
		// TLS 类型
		if certFile, certData, err := handleFormFile(r, "cert", tempDir, name, "", "crt"); err != nil {
			BadRequest(w, err.Error())
			return
		} else if certFile != "" {
			tempFiles["cert"] = certFile
			finalFiles["cert"] = filepath.Join(keystoreDir, name+".crt")

			// 如果同时上传了密钥，验证配对
			if keyFile, keyData, err := handleFormFile(r, "key", tempDir, name, "", "key"); err != nil {
				BadRequest(w, err.Error())
				return
			} else if keyFile != "" {
				tempFiles["key"] = keyFile
				finalFiles["key"] = filepath.Join(keystoreDir, name+".key")

				if err := keystore.VerifyCertificateKeyPair(certData, keyData, false); err != nil {
					BadRequest(w, "证书与密钥不匹配: "+err.Error())
					return
				}
			}
		} else if keyFile, _, err := handleFormFile(r, "key", tempDir, name, "", "key"); err != nil {
			BadRequest(w, err.Error())
			return
		} else if keyFile != "" {
			BadRequest(w, "上传密钥时必须同时上传证书")
			return
		}
	}

	if len(tempFiles) == 0 {
		BadRequest(w, "请至少上传一个文件")
		return
	}

	// 原子替换文件
	for tempPath, finalPath := range finalFiles {
		if err := os.Rename(tempPath, finalPath); err != nil {
			InternalError(w, "文件替换失败: "+err.Error())
			return
		}

		// 设置文件权限
		if strings.HasSuffix(finalPath, ".key") {
			os.Chmod(finalPath, 0600)
		} else {
			os.Chmod(finalPath, 0644)
		}
	}

	// 更新配置文件中的参数
	for i := range c.cfg.KeyStores {
		if c.cfg.KeyStores[i].Name == name {
			if c.cfg.KeyStores[i].Params == nil {
				c.cfg.KeyStores[i].Params = make(map[string]string)
			}
			for key, finalPath := range finalFiles {
				// 转换为相对路径
				relPath, err := filepath.Rel(c.cfg.WorkDir, finalPath)
				if err == nil {
					c.cfg.KeyStores[i].Params[key] = relPath
				}
			}
			break
		}
	}

	// 保存配置文件
	if err := config.Save(c.cfg); err != nil {
		InternalError(w, "保存配置失败: "+err.Error())
		return
	}

	// 重新加载 keystore
	updatedInfo, err := c.keyStoreMgr.Get(name)
	if err != nil {
		InternalError(w, "重新加载 keystore 失败: "+err.Error())
		return
	}

	c.log.Info("更新 keystore 证书和密钥: %s", name)
	Success(w, updatedInfo)
}

// handleFormFile 处理表单文件上传
// 参数：
//   - r: HTTP 请求
//   - fieldName: 表单字段名
//   - tempDir: 临时目录
//   - name: keystore 名称
//   - prefix: 文件前缀（sign/enc）
//   - ext: 文件扩展名（crt/key）
//
// 返回：
//   - string: 临时文件路径，如果没有上传文件则为空字符串
//   - []byte: 文件内容
//   - error: 错误信息
func handleFormFile(r *http.Request, fieldName, tempDir, name, prefix, ext string) (string, []byte, error) {
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil, nil
		}
		return "", nil, fmt.Errorf("读取 %s 文件失败: %w", fieldName, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", nil, fmt.Errorf("读取 %s 文件内容失败: %w", fieldName, err)
	}

	// 生成临时文件名
	tempFileName := name + "-" + prefix + "." + ext
	if prefix == "" {
		tempFileName = name + "." + ext
	}
	tempPath := filepath.Join(tempDir, tempFileName+"."+header.Filename)

	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return "", nil, fmt.Errorf("写入临时文件失败: %w", err)
	}

	return tempPath, data, nil
}
