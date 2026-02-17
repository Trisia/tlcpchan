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

// ListKeyStores 列出所有 keystore
func (c *SecurityController) ListKeyStores(w http.ResponseWriter, r *http.Request) {
	keyStores := c.keyStoreMgr.List()
	Success(w, keyStores)
}

// CreateKeyStore 创建 keystore
// 该方法会：
// 1. 在 keystore 管理器中创建 keystore
// 2. 更新配置文件中的 keystores 列表
// 3. 保存配置文件到磁盘
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
// 该方法会：
// 1. 从 keystore 管理器中删除 keystore
// 2. 从配置文件的 keystores 列表中移除
// 3. 保存更新后的配置文件到磁盘
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

// GetRootCert 获取根证书
func (c *SecurityController) GetRootCert(w http.ResponseWriter, r *http.Request) {
	filename := PathParam(r, "filename")
	cert, err := c.rootCertMgr.Get(filename)
	if err != nil {
		NotFound(w, "根证书不存在")
		return
	}
	Success(w, cert)
}

// DeleteRootCert 删除根证书
func (c *SecurityController) DeleteRootCert(w http.ResponseWriter, r *http.Request) {
	filename := PathParam(r, "filename")
	if err := c.rootCertMgr.Delete(filename); err != nil {
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
