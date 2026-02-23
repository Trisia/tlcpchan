package controller

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/security/certgen"
	"github.com/Trisia/tlcpchan/security/keystore"
)

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
 * @apiSuccess {String} -.notBefore 证书生效时间，ISO 8601 格式
 * @apiSuccess {String} -.notAfter 证书过期时间，ISO 8601 格式
 * @apiSuccess {String} -.keyType 密钥类型（如 "SM2", "RSA-2048", "ECDSA-P256"）
 * @apiSuccess {String} -.serialNumber 证书序列号（十六进制）
 * @apiSuccess {Number} -.version 证书版本
 * @apiSuccess {Boolean} -.isCA 是否为 CA 证书
 * @apiSuccess {String[]} -.keyUsage 密钥用途数组
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     [
 *       {
 *         "filename": "root-ca.crt",
 *         "subject": "CN=Root CA, O=Example Org",
 *         "issuer": "CN=Root CA, O=Example Org",
 *         "notBefore": "2020-01-01T00:00:00Z",
 *         "notAfter": "2030-01-01T00:00:00Z",
 *         "keyType": "SM2",
 *         "serialNumber": "0102030405060708090a1b1c1d1e1f1011121314",
 *         "version": 3,
 *         "isCA": true,
 *         "keyUsage": ["Cert Sign", "CRL Sign"]
 *       }
 *     ]
 */
func (c *SecurityController) ListRootCerts(w http.ResponseWriter, r *http.Request) {
	certs := c.rootCertMgr.List()
	Success(w, certs)
}

/**
 * @api {get} /api/security/rootcerts/:filename 下载根证书
 * @apiName GetRootCert
 * @apiGroup Security-RootCert
 * @apiVersion 1.0.0
 *
 * @apiDescription 下载指定根证书文件
 *
 * @apiParam {String} filename 证书文件名（路径参数），唯一标识符
 *
 * @apiSuccess {File} - 证书文件内容（PEM 格式）
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     根证书不存在
 */
func (c *SecurityController) GetRootCert(w http.ResponseWriter, r *http.Request) {
	filename := PathParam(r, "filename")
	certData, err := c.rootCertMgr.ReadFile(filename)
	if err != nil {
		NotFound(w, "根证书不存在")
		return
	}

	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Write(certData)
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
	c.log.Info("删除根证书: %s", filename)
	Success(w, nil)
}

/**
 * @api {post} /api/security/rootcerts 添加根证书
 * @apiName AddRootCert
 * @apiGroup Security-RootCert
 * @apiVersion 1.0.0
 *
 * @apiDescription 上传并添加新的根证书到系统中
 *
 * @apiBody {String} filename 文件名，表单字段，用于指定保存的证书文件名（必需）
 * @apiBody {File} cert 证书文件，表单字段，PEM 或 DER 格式的证书文件，最大 10MB
 *
 * @apiSuccess {String} filename 证书文件名
 * @apiSuccess {String} subject 证书主题（Subject）
 * @apiSuccess {String} issuer 证书颁发者（Issuer）
 * @apiSuccess {String} notBefore 证书生效时间，ISO 8601 格式
 * @apiSuccess {String} notAfter 证书过期时间，ISO 8601 格式
 * @apiSuccess {String} keyType 密钥类型（如 "SM2", "RSA-2048", "ECDSA-P256"）
 * @apiSuccess {String} serialNumber 证书序列号（十六进制）
 * @apiSuccess {Number} version 证书版本
 * @apiSuccess {Boolean} isCA 是否为 CA 证书
 * @apiSuccess {String[]} keyUsage 密钥用途数组
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "filename": "root-ca.crt",
 *       "subject": "CN=Root CA, O=Example Org",
 *       "issuer": "CN=Root CA, O=Example Org",
 *       "notBefore": "2020-01-01T00:00:00Z",
 *       "notAfter": "2030-01-01T00:00:00Z",
 *       "keyType": "SM2",
 *       "serialNumber": "0102030405060708090a1b1c1d1e1f1011121314",
 *       "version": 3,
 *       "isCA": true,
 *       "keyUsage": ["Cert Sign", "CRL Sign"]
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

	c.log.Info("添加根证书: %s", filename)
	Success(w, cert)
}

/**
 * @api {post} /api/security/rootcerts/reload 重载根证书
 * @apiName ReloadRootCerts
 * @apiGroup Security-RootCert
 * @apiVersion 1.0.0
 *
 * @apiDescription 重新加载所有根证书
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
	c.log.Info("重载根证书")
	Success(w, nil)
}

/**
 * @api {post} /api/security/rootcerts/generate 生成根 CA 证书
 * @apiName GenerateRootCA
 * @apiGroup Security-RootCert
 * @apiVersion 1.0.0
 *
 * @apiDescription 生成新的自签名根 CA 证书，并添加到信任证书列表中，同时创建对应的 keystore
 *
 * @apiBody {String} [type="tlcp"] 证书类型，可选值：tlcp（SM2）、tls（RSA-2048）
 * @apiBody {String} [commonName="tlcpchan-root-ca"] 证书通用名称（CN）
 * @apiBody {String} [country] 国家代码（C），如 "CN"
 * @apiBody {String} [stateOrProvince] 省/州（ST），如 "Beijing"
 * @apiBody {String} [locality] 地区（L），如 "Haidian"
 * @apiBody {String} [org="tlcpchan"] 组织名称（O）
 * @apiBody {String} [orgUnit] 组织单位（OU），如 "IT"
 * @apiBody {String} [emailAddress] 邮箱地址
 * @apiBody {Number} [years=10] 证书有效期（年）
 * @apiBody {Number} [days] 证书有效期（天），与 years 二选一
 *
 * @apiSuccess {String} filename 证书文件名
 * @apiSuccess {String} subject 证书主题（Subject）
 * @apiSuccess {String} issuer 证书颁发者（Issuer）
 * @apiSuccess {String} notBefore 证书生效时间，ISO 8601 格式
 * @apiSuccess {String} notAfter 证书过期时间，ISO 8601 格式
 * @apiSuccess {String} keyType 密钥类型（如 "SM2", "RSA-2048", "ECDSA-P256"）
 * @apiSuccess {String} serialNumber 证书序列号（十六进制）
 * @apiSuccess {Number} version 证书版本
 * @apiSuccess {Boolean} isCA 是否为 CA 证书
 * @apiSuccess {String[]} keyUsage 密钥用途数组
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "filename": "my-root-ca.crt",
 *       "subject": "CN=my-root-ca, O=My Org",
 *       "issuer": "CN=my-root-ca, O=My Org",
 *       "notBefore": "2024-01-01T00:00:00Z",
 *       "notAfter": "2034-01-01T00:00:00Z",
 *       "keyType": "SM2",
 *       "serialNumber": "0102030405060708090a1b1c1d1e1f1011121314",
 *       "version": 3,
 *       "isCA": true,
 *       "keyUsage": ["Cert Sign", "CRL Sign"]
 *     }
 *
 * @apiParamExample {json} Request-Example:
 *     {
 *       "type": "tlcp",
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

	c.log.Info("生成根CA证书: %s", req.CommonName)
	Success(w, cert)
}
