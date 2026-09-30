package security

import (
	"github.com/Trisia/tlcpchan/security/ibcparams"
	"github.com/Trisia/tlcpchan/security/keystore"
	"github.com/Trisia/tlcpchan/security/rootcert"
)

type (
	KeyStore        = keystore.KeyStore
	KeyStoreType    = keystore.KeyStoreType
	LoaderType      = keystore.LoaderType
	KeyStoreInfo    = keystore.KeyStoreInfo
	KeyStoreManager = keystore.Manager
	KeyType         = keystore.KeyType
	IBCKeyStore     = keystore.IBCKeyStore
	IBCInfo         = keystore.IBCInfo
	RootCert        = rootcert.RootCert
	RootCertPool    = rootcert.RootCertPool
	RootCertManager = rootcert.Manager
	IBCParam        = ibcparams.IBCParam
	IBCParamManager = ibcparams.Manager
)

const (
	KeyStoreTypeTLCP  = keystore.KeyStoreTypeTLCP
	KeyStoreTypeTLS   = keystore.KeyStoreTypeTLS
	KeyStoreTypeIBC   = keystore.KeyStoreTypeIBC
	LoaderTypeFile    = keystore.LoaderTypeFile
	LoaderTypeNamed   = keystore.LoaderTypeNamed
	LoaderTypeSKF     = keystore.LoaderTypeSKF
	LoaderTypeSDF     = keystore.LoaderTypeSDF
	LoaderTypeIBCFile = keystore.LoaderTypeIBCFile
	KeyTypeSign       = keystore.KeyTypeSign
	KeyTypeEnc        = keystore.KeyTypeEnc
)

func NewKeyStoreManager() *KeyStoreManager {
	return keystore.NewManager()
}

func NewRootCertManager(baseDir string) *RootCertManager {
	return rootcert.NewManager(baseDir)
}

// NewIBCParamManager 创建 IBC 信任池管理器。
//
// 参数：
//   - baseDir: 信任池目录，通常为 <workDir>/ibcparams
//
// 返回值：
//   - *IBCParamManager: 信任池管理器实例
func NewIBCParamManager(baseDir string) *IBCParamManager {
	return ibcparams.NewManager(baseDir)
}
