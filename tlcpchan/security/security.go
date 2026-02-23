package security

import (
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
	RootCert        = rootcert.RootCert
	RootCertPool    = rootcert.RootCertPool
	RootCertManager = rootcert.Manager
)

const (
	KeyStoreTypeTLCP = keystore.KeyStoreTypeTLCP
	KeyStoreTypeTLS  = keystore.KeyStoreTypeTLS
	LoaderTypeFile   = keystore.LoaderTypeFile
	LoaderTypeNamed  = keystore.LoaderTypeNamed
	LoaderTypeSKF    = keystore.LoaderTypeSKF
	LoaderTypeSDF    = keystore.LoaderTypeSDF
	KeyTypeSign      = keystore.KeyTypeSign
	KeyTypeEnc       = keystore.KeyTypeEnc
)

func NewKeyStoreManager() *KeyStoreManager {
	return keystore.NewManager()
}

func NewRootCertManager(baseDir string) *RootCertManager {
	return rootcert.NewManager(baseDir)
}
