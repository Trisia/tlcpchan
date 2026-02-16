package proxy

import (
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/security"
)

type HTTPServerProxy struct {
	*ServerProxy
}

func NewHTTPServerProxy(cfg *config.InstanceConfig,
	keyStoreMgr *security.KeyStoreManager,
	rootCertMgr *security.RootCertManager) (*HTTPServerProxy, error) {
	sp, err := NewServerProxy(cfg, keyStoreMgr, rootCertMgr)
	if err != nil {
		return nil, err
	}
	return &HTTPServerProxy{ServerProxy: sp}, nil
}
