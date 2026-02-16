package proxy

import (
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/security"
)

type HTTPClientProxy struct {
	*ClientProxy
}

func NewHTTPClientProxy(cfg *config.InstanceConfig,
	keyStoreMgr *security.KeyStoreManager,
	rootCertMgr *security.RootCertManager) (*HTTPClientProxy, error) {
	cp, err := NewClientProxy(cfg, keyStoreMgr, rootCertMgr)
	if err != nil {
		return nil, err
	}
	return &HTTPClientProxy{ClientProxy: cp}, nil
}
