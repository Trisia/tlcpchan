package controller

import (
	"encoding/json"
	"net/http"

	"github.com/Trisia/tlcpchan/security"
	"github.com/Trisia/tlcpchan/security/keystore"
)

// SecurityController 安全参数管理控制器
type SecurityController struct {
	keyStoreMgr *security.KeyStoreManager
	rootCertMgr *security.RootCertManager
}

// NewSecurityController 创建安全参数管理控制器
func NewSecurityController(keyStoreMgr *security.KeyStoreManager, rootCertMgr *security.RootCertManager) *SecurityController {
	return &SecurityController{
		keyStoreMgr: keyStoreMgr,
		rootCertMgr: rootCertMgr,
	}
}

// RegisterRoutes 注册路由
func (c *SecurityController) RegisterRoutes(r *Router) {
	r.GET("/api/v1/security/keystores", c.ListKeyStores)
	r.POST("/api/v1/security/keystores", c.CreateKeyStore)
	r.GET("/api/v1/security/keystores/:name", c.GetKeyStore)
	r.DELETE("/api/v1/security/keystores/:name", c.DeleteKeyStore)
	r.POST("/api/v1/security/keystores/:name/reload", c.ReloadKeyStore)

	r.GET("/api/v1/security/rootcerts", c.ListRootCerts)
	r.POST("/api/v1/security/rootcerts", c.AddRootCert)
	r.GET("/api/v1/security/rootcerts/:name", c.GetRootCert)
	r.DELETE("/api/v1/security/rootcerts/:name", c.DeleteRootCert)
	r.POST("/api/v1/security/rootcerts/reload", c.ReloadRootCerts)
}

// ListKeyStores 列出所有 keystore
func (c *SecurityController) ListKeyStores(w http.ResponseWriter, r *http.Request) {
	keyStores := c.keyStoreMgr.List()
	Success(w, keyStores)
}

// CreateKeyStore 创建 keystore
func (c *SecurityController) CreateKeyStore(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string                `json:"name"`
		LoaderType   keystore.LoaderType   `json:"loaderType"`
		LoaderConfig keystore.LoaderConfig `json:"loaderConfig"`
		Protected    bool                  `json:"protected"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "无效的请求: "+err.Error())
		return
	}

	if req.Name == "" {
		BadRequest(w, "名称不能为空")
		return
	}

	info, err := c.keyStoreMgr.Create(req.Name, req.LoaderConfig, req.Protected)
	if err != nil {
		InternalError(w, "创建失败: "+err.Error())
		return
	}

	Success(w, info)
}

// GetKeyStore 获取 keystore
func (c *SecurityController) GetKeyStore(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	info, err := c.keyStoreMgr.Get(name)
	if err != nil {
		NotFound(w, "keystore 不存在")
		return
	}
	Success(w, info)
}

// DeleteKeyStore 删除 keystore
func (c *SecurityController) DeleteKeyStore(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	if err := c.keyStoreMgr.Delete(name); err != nil {
		InternalError(w, "删除失败: "+err.Error())
		return
	}
	Success(w, nil)
}

// ReloadKeyStore 重载 keystore
func (c *SecurityController) ReloadKeyStore(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	if err := c.keyStoreMgr.Reload(name); err != nil {
		InternalError(w, "重载失败: "+err.Error())
		return
	}
	Success(w, nil)
}

// ListRootCerts 列出所有根证书
func (c *SecurityController) ListRootCerts(w http.ResponseWriter, r *http.Request) {
	certs := c.rootCertMgr.List()
	Success(w, certs)
}

// AddRootCert 添加根证书
func (c *SecurityController) AddRootCert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		BadRequest(w, "解析表单失败: "+err.Error())
		return
	}

	name := r.FormValue("name")
	if name == "" {
		BadRequest(w, "名称不能为空")
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

	cert, err := c.rootCertMgr.Add(name, certData)
	if err != nil {
		InternalError(w, "添加失败: "+err.Error())
		return
	}

	Success(w, cert)
}

// GetRootCert 获取根证书
func (c *SecurityController) GetRootCert(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	cert, err := c.rootCertMgr.Get(name)
	if err != nil {
		NotFound(w, "根证书不存在")
		return
	}
	Success(w, cert)
}

// DeleteRootCert 删除根证书
func (c *SecurityController) DeleteRootCert(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	if err := c.rootCertMgr.Delete(name); err != nil {
		InternalError(w, "删除失败: "+err.Error())
		return
	}
	Success(w, nil)
}

// ReloadRootCerts 重载所有根证书
func (c *SecurityController) ReloadRootCerts(w http.ResponseWriter, r *http.Request) {
	if err := c.rootCertMgr.Reload(); err != nil {
		InternalError(w, "重载失败: "+err.Error())
		return
	}
	Success(w, nil)
}
