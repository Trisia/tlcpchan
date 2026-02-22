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

// CSRParams 证书请求参数
type CSRParams struct {
	CommonName      string   `json:"commonName"`
	Country         string   `json:"country,omitempty"`
	StateOrProvince string   `json:"stateOrProvince,omitempty"`
	Locality        string   `json:"locality,omitempty"`
	Org             string   `json:"org,omitempty"`
	OrgUnit         string   `json:"orgUnit,omitempty"`
	EmailAddress    string   `json:"emailAddress,omitempty"`
	DNSNames        []string `json:"dnsNames,omitempty"`
	IPAddresses     []string `json:"ipAddresses,omitempty"`
}

// KeyStore 抽象 keystore 接口
type KeyStore interface {
	Type() KeyStoreType
	TLCPCertificate() ([]*tlcp.Certificate, error)
	TLSCertificate() (*tls.Certificate, error)
	GenerateCSR(keyType KeyType, params CSRParams) ([]byte, error)
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
