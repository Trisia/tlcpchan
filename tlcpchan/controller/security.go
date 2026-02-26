package controller

import (
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/security"
)

// SecurityController 安全参数管理控制器
// 负责管理 keystore 和根证书，同时负责更新配置文件
type SecurityController struct {
	keyStoreMgr *security.KeyStoreManager // keystore 管理器
	rootCertMgr *security.RootCertManager // 根证书管理器
	cfg         *config.Config            // 全局配置
	configPath  string                    // 配置文件路径
	log         *logger.Logger            // 日志记录器
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
		log:         logger.Default(),
	}
}

// RegisterRoutes 注册路由
func (c *SecurityController) RegisterRoutes(r *Router) {
	r.GET("/api/security/keystores", c.ListKeyStores)
	r.POST("/api/security/keystores", c.CreateKeyStore)
	r.POST("/api/security/keystores/generate", c.GenerateKeyStore)
	r.GET("/api/security/keystores/:name", c.GetKeyStore)
	r.PUT("/api/security/keystores/:name", c.UpdateKeystoreParams)
	r.POST("/api/security/keystores/:name/upload", c.UpdateCertificates)
	r.GET("/api/security/keystores/:name/instances", c.GetKeyStoreInstances)
	r.DELETE("/api/security/keystores/:name", c.DeleteKeyStore)
	r.POST("/api/security/keystores/:name/export-csr", c.ExportCSR)

	r.GET("/api/security/rootcerts", c.ListRootCerts)
	r.POST("/api/security/rootcerts", c.AddRootCert)
	r.POST("/api/security/rootcerts/generate", c.GenerateRootCA)
	r.GET("/api/security/rootcerts/:filename", c.GetRootCert)
	r.DELETE("/api/security/rootcerts/:filename", c.DeleteRootCert)
	r.POST("/api/security/rootcerts/reload", c.ReloadRootCerts)
}
