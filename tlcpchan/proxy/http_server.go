package proxy

import (
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/security"
)

type HTTPServerProxy struct {
	*ServerProxy
}

// NewHTTPServerProxy 创建 HTTP 服务端代理。
//
// 参数：
//   - cfg: 实例配置
//   - keyStoreMgr: keystore 管理器
//   - rootCertMgr: 根证书管理器
//   - ibcParamMgr: IBC 信任池管理器，可为 nil（按空信任池处理）
//
// 返回值：
//   - *HTTPServerProxy: HTTP 服务端代理实例
//   - error: 底层 TCP 服务端代理创建失败时返回错误
func NewHTTPServerProxy(cfg *config.InstanceConfig,
	keyStoreMgr *security.KeyStoreManager,
	rootCertMgr *security.RootCertManager,
	ibcParamMgr *security.IBCParamManager) (*HTTPServerProxy, error) {
	sp, err := NewServerProxy(cfg, keyStoreMgr, rootCertMgr, ibcParamMgr)
	if err != nil {
		return nil, err
	}
	return &HTTPServerProxy{ServerProxy: sp}, nil
}
