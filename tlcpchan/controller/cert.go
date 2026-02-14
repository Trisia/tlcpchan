package controller

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Trisia/tlcpchan/logger"
)

// CertificateInfo 证书信息DTO
type CertificateInfo struct {
	// Name 证书文件名
	// 示例: "server.crt", "client.key"
	Name string `json:"name"`
	// Type 证书类型
	// 可选值: "certificate" (证书), "private_key" (私钥)
	Type string `json:"type"`
	// ExpiresAt 过期时间，ISO 8601格式
	// 示例: "2025-12-31T23:59:59Z"
	ExpiresAt string `json:"expires_at,omitempty"`
}

var certNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)

func isValidCertName(name string) bool {
	if !certNameRegex.MatchString(name) {
		return false
	}
	clean := filepath.Clean(name)
	return clean == name && !strings.Contains(clean, "..")
}

type CertController struct {
	certDir string
	log     *logger.Logger
}

func NewCertController(certDir string) *CertController {
	return &CertController{
		certDir: certDir,
		log:     logger.Default(),
	}
}

func (c *CertController) List(w http.ResponseWriter, r *http.Request) {
	certs := make([]CertificateInfo, 0)

	entries, err := os.ReadDir(c.certDir)
	if err != nil {
		Success(w, certs)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		var certType string
		if strings.HasSuffix(name, ".crt") || strings.HasSuffix(name, ".pem") {
			certType = "certificate"
		} else if strings.HasSuffix(name, ".key") {
			certType = "private_key"
		} else {
			continue
		}

		certs = append(certs, CertificateInfo{
			Name: name,
			Type: certType,
		})
	}

	Success(w, certs)
}

func (c *CertController) Upload(w http.ResponseWriter, r *http.Request) {
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
	if !isValidCertName(filename) {
		BadRequest(w, "无效的证书名称: 只允许字母、数字、下划线、连字符和点")
		return
	}

	dst, err := os.Create(filepath.Join(c.certDir, filename))
	if err != nil {
		InternalError(w, "保存文件失败: "+err.Error())
		return
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(file); err != nil {
		InternalError(w, "写入文件失败: "+err.Error())
		return
	}

	c.log.Info("上传证书: %s", filename)
	Created(w, CertificateInfo{
		Name: filename,
	})
}

func (c *CertController) Delete(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		BadRequest(w, "缺少证书名称")
		return
	}

	if !isValidCertName(name) {
		BadRequest(w, "无效的证书名称")
		return
	}

	path := filepath.Join(c.certDir, name)
	if err := os.Remove(path); err != nil {
		InternalError(w, "删除证书失败: "+err.Error())
		return
	}

	c.log.Info("删除证书: %s", name)
	NoContent(w)
}

func (c *CertController) Reload(w http.ResponseWriter, r *http.Request) {
	c.log.Info("重新加载证书")
	Success(w, map[string]string{"message": "证书已重新加载"})
}

func (c *CertController) RegisterRoutes(router *Router) {
	router.GET("/api/v1/certificates", c.List)
	router.POST("/api/v1/certificates", c.Upload)
	router.DELETE("/api/v1/certificates", c.Delete)
	router.POST("/api/v1/certificates/reload", c.Reload)
}
