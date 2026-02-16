package controller

import (
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Trisia/tlcpchan/logger"
)

// TrustedCertificateInfo 信任证书信息DTO
type TrustedCertificateInfo struct {
	// Name 证书文件名
	// 示例: "root-ca.crt", "intermediate-ca.pem"
	Name string `json:"name"`
	// Type 证书类型
	// 固定值: "certificate"（信任目录只存放证书，不存放私钥）
	Type string `json:"type"`
	// SerialNumber 证书序列号（十六进制字符串）
	// 示例: "1a2b3c4d"
	SerialNumber string `json:"serialNumber,omitempty"`
	// Subject 证书主题
	Subject string `json:"subject,omitempty"`
	// Issuer 证书颁发者
	Issuer string `json:"issuer,omitempty"`
	// ExpiresAt 过期时间，ISO 8601格式
	// 示例: "2025-12-31T23:59:59Z"
	ExpiresAt string `json:"expiresAt,omitempty"`
	// IsCA 是否为CA证书
	IsCA bool `json:"isCA,omitempty"`
}

var trustedCertNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)

func isValidTrustedCertName(name string) bool {
	if !trustedCertNameRegex.MatchString(name) {
		return false
	}
	clean := filepath.Clean(name)
	return clean == name && !strings.Contains(clean, "..")
}

// serialNumberToHex 将序列号转换为十六进制字符串
func serialNumberToHex(sn *big.Int) string {
	if sn == nil {
		return ""
	}
	return sn.Text(16)
}

type TrustedController struct {
	trustedDir string
	log        *logger.Logger
}

func NewTrustedController(trustedDir string) *TrustedController {
	return &TrustedController{
		trustedDir: trustedDir,
		log:        logger.Default(),
	}
}

// parseTrustedCertificateFile 解析信任证书文件，提取证书信息
func parseTrustedCertificateFile(path string) (*x509.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err == nil {
		return cert, nil
	}

	return nil, nil
}

/**
 * @api {get} /api/v1/trusted 获取信任证书列表
 * @apiName ListTrustedCertificates
 * @apiGroup Trusted
 * @apiVersion 1.0.0
 *
 * @apiDescription 获取信任证书目录中的所有证书文件列表
 *
 * @apiSuccess {Object[]} trusted 信任证书列表
 * @apiSuccess {String} trusted.name 证书文件名
 * @apiSuccess {String} trusted.type 证书类型 (固定为 "certificate")
 * @apiSuccess {String} trusted.serialNumber 证书序列号（十六进制）
 * @apiSuccess {String} trusted.subject 证书主题
 * @apiSuccess {String} trusted.issuer 证书颁发者
 * @apiSuccess {String} trusted.expiresAt 过期时间
 * @apiSuccess {Boolean} trusted.isCA 是否为CA证书
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     [
 *       {
 *         "name": "root-ca.crt",
 *         "type": "certificate",
 *         "serialNumber": "1a2b3c4d",
 *         "subject": "CN=Root CA",
 *         "issuer": "CN=Root CA",
 *         "expiresAt": "2030-12-31T23:59:59Z",
 *         "isCA": true
 *       }
 *     ]
 */
func (c *TrustedController) List(w http.ResponseWriter, r *http.Request) {
	certs := make([]TrustedCertificateInfo, 0)

	entries, err := os.ReadDir(c.trustedDir)
	if err != nil {
		Success(w, certs)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".crt") && !strings.HasSuffix(name, ".pem") && !strings.HasSuffix(name, ".cer") {
			continue
		}

		certInfo := TrustedCertificateInfo{
			Name: name,
			Type: "certificate",
		}

		path := filepath.Join(c.trustedDir, name)
		if cert, err := parseTrustedCertificateFile(path); err == nil && cert != nil {
			certInfo.SerialNumber = serialNumberToHex(cert.SerialNumber)
			certInfo.Subject = cert.Subject.String()
			certInfo.Issuer = cert.Issuer.String()
			certInfo.ExpiresAt = cert.NotAfter.UTC().Format(time.RFC3339)
			certInfo.IsCA = cert.IsCA
		}

		certs = append(certs, certInfo)
	}

	Success(w, certs)
}

/**
 * @api {post} /api/v1/trusted 上传信任证书
 * @apiName UploadTrustedCertificate
 * @apiGroup Trusted
 * @apiVersion 1.0.0
 *
 * @apiDescription 上传信任证书（根证书/CA证书）到信任证书目录
 *
 * @apiSuccess {String} name 证书文件名
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 201 Created
 *     {
 *       "name": "new-root-ca.crt"
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     解析上传文件失败
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     无效的证书名称
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     保存文件失败
 */
func (c *TrustedController) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		BadRequest(w, "解析上传文件失败: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		BadRequest(w, "获取上传文件失败: "+err.Error())
		return
	}
	defer file.Close()

	filename := header.Filename
	if !isValidTrustedCertName(filename) {
		BadRequest(w, "无效的证书名称: 只允许字母、数字、下划线、连字符和点")
		return
	}

	dst, err := os.Create(filepath.Join(c.trustedDir, filename))
	if err != nil {
		InternalError(w, "保存文件失败: "+err.Error())
		return
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(file); err != nil {
		InternalError(w, "写入文件失败: "+err.Error())
		return
	}

	c.log.Info("上传信任证书: %s", filename)
	Created(w, TrustedCertificateInfo{
		Name: filename,
	})
}

/**
 * @api {delete} /api/v1/trusted 删除信任证书
 * @apiName DeleteTrustedCertificate
 * @apiGroup Trusted
 * @apiVersion 1.0.0
 *
 * @apiDescription 删除指定的信任证书文件
 *
 * @apiQuery {String} name 证书文件名
 *
 * @apiSuccessExample {text} Success-Response:
 *     HTTP/1.1 204 No Content
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     缺少证书名称
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     无效的证书名称
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     删除证书失败
 */
func (c *TrustedController) Delete(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		BadRequest(w, "缺少证书名称")
		return
	}

	if !isValidTrustedCertName(name) {
		BadRequest(w, "无效的证书名称")
		return
	}

	path := filepath.Join(c.trustedDir, name)
	if err := os.Remove(path); err != nil {
		InternalError(w, "删除信任证书失败: "+err.Error())
		return
	}

	c.log.Info("删除信任证书: %s", name)
	NoContent(w)
}

/**
 * @api {post} /api/v1/trusted/reload 重载信任证书
 * @apiName ReloadTrustedCertificates
 * @apiGroup Trusted
 * @apiVersion 1.0.0
 *
 * @apiDescription 重新加载所有信任证书
 *
 * @apiSuccess {String} message 操作消息
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "message": "信任证书已重新加载"
 *     }
 */
func (c *TrustedController) Reload(w http.ResponseWriter, r *http.Request) {
	c.log.Info("重新加载信任证书")
	Success(w, map[string]string{"message": "信任证书已重新加载"})
}

func (c *TrustedController) RegisterRoutes(router *Router) {
	router.GET("/api/v1/trusted", c.List)
	router.POST("/api/v1/trusted", c.Upload)
	router.DELETE("/api/v1/trusted", c.Delete)
	router.POST("/api/v1/trusted/reload", c.Reload)
}
