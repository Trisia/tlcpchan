package keystore

import (
	"crypto/tls"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
)

// KeyStoreType keystore 类型
type KeyStoreType string

const (
	KeyStoreTypeTLCP KeyStoreType = "tlcp"
	KeyStoreTypeTLS  KeyStoreType = "tls"
)

// KeyType 密钥类型
type KeyType string

const (
	KeyTypeSign KeyType = "sign"
	KeyTypeEnc  KeyType = "enc"
)

// KeyStore 抽象 keystore 接口
type KeyStore interface {
	Type() KeyStoreType
	TLCPCertificate() ([]*tlcp.Certificate, error)
	TLSCertificate() (*tls.Certificate, error)
}

// LoaderType 加载器类型
type LoaderType string

const (
	LoaderTypeFile  LoaderType = "file"
	LoaderTypeNamed LoaderType = "named"
	LoaderTypeSKF   LoaderType = "skf"
	LoaderTypeSDF   LoaderType = "sdf"
)

// KeyStoreInfo keystore 信息
type KeyStoreInfo struct {
	Name       string            `json:"name" yaml:"name"`
	Type       KeyStoreType      `json:"type" yaml:"type"`
	LoaderType LoaderType        `json:"loaderType" yaml:"loaderType"`
	Params     map[string]string `json:"params" yaml:"params"`
	Protected  bool              `json:"protected" yaml:"protected"`
	CreatedAt  time.Time         `json:"createdAt" yaml:"createdAt"`
	UpdatedAt  time.Time         `json:"updatedAt" yaml:"updatedAt"`
}
