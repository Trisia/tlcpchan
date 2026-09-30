package proxy

import (
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/security"
)

type HTTPClientProxy struct {
	*ClientProxy
}

// NewHTTPClientProxy 创建 HTTP 客户端代理。
//
// 参数：
//   - cfg: 实例配置
//   - keyStoreMgr: keystore 管理器
//   - rootCertMgr: 根证书管理器
//   - ibcParamMgr: IBC 信任池管理器，可为 nil（按空信任池处理）
//
// 返回值：
//   - *HTTPClientProxy: HTTP 客户端代理实例
//   - error: 底层 TCP 客户端代理创建失败时返回错误
func NewHTTPClientProxy(cfg *config.InstanceConfig,
	keyStoreMgr *security.KeyStoreManager,
	rootCertMgr *security.RootCertManager,
	ibcParamMgr *security.IBCParamManager) (*HTTPClientProxy, error) {
	cp, err := NewClientProxy(cfg, keyStoreMgr, rootCertMgr, ibcParamMgr)
	if err != nil {
		return nil, err
	}
	return &HTTPClientProxy{ClientProxy: cp}, nil
}
