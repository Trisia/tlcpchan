package security

import (
	"github.com/Trisia/tlcpchan/security/keystore"
	"github.com/Trisia/tlcpchan/security/rootcert"
)

type (
	KeyStore        = keystore.KeyStore
	KeyStoreType    = keystore.KeyStoreType
	LoaderType      = keystore.LoaderType
	LoaderConfig    = keystore.LoaderConfig
	KeyStoreInfo    = keystore.KeyStoreInfo
	KeyStoreManager = keystore.Manager
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
)

func NewKeyStoreManager(baseDir string) *KeyStoreManager {
	return keystore.NewManager(baseDir)
}

func NewRootCertManager(baseDir string) *RootCertManager {
	return rootcert.NewManager(baseDir)
}
