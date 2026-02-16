package key

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/emmansun/gmsm/smx509"
)

// Validator 证书密钥验证器
type Validator struct{}

// NewValidator 创建验证器
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateCreateTLCP 验证国密密钥创建
func (v *Validator) ValidateCreateTLCP(signCert, signKey, encCert, encKey []byte) error {
	if len(signCert) == 0 {
		return fmt.Errorf("签名证书不能为空")
	}
	if len(signKey) == 0 {
		return fmt.Errorf("签名密钥不能为空")
	}
	if len(encCert) == 0 {
		return fmt.Errorf("加密证书不能为空")
	}
	if len(encKey) == 0 {
		return fmt.Errorf("加密密钥不能为空")
	}

	if err := v.ValidateCertKeyPair(signCert, signKey, true); err != nil {
		return fmt.Errorf("签名证书与密钥不匹配: %w", err)
	}

	if err := v.ValidateCertKeyPair(encCert, encKey, true); err != nil {
		return fmt.Errorf("加密证书与密钥不匹配: %w", err)
	}

	return nil
}

// ValidateCreateTLS 验证国际密钥创建
func (v *Validator) ValidateCreateTLS(signCert, signKey []byte) error {
	if len(signCert) == 0 {
		return fmt.Errorf("签名证书不能为空")
	}
	if len(signKey) == 0 {
		return fmt.Errorf("签名密钥不能为空")
	}

	if err := v.ValidateCertKeyPair(signCert, signKey, false); err != nil {
		return fmt.Errorf("证书与密钥不匹配: %w", err)
	}

	return nil
}

// ValidateUpdateCertsTLCP 验证国密证书更新
func (v *Validator) ValidateUpdateCertsTLCP(signCert, encCert []byte) error {
	if len(signCert) == 0 && len(encCert) == 0 {
		return fmt.Errorf("至少需要提供一个证书")
	}
	return nil
}

// ValidateUpdateCertsTLS 验证国际证书更新
func (v *Validator) ValidateUpdateCertsTLS(signCert []byte) error {
	if len(signCert) == 0 {
		return fmt.Errorf("签名证书不能为空")
	}
	return nil
}

// ValidateCertKeyPair 验证证书和密钥是否匹配
func (v *Validator) ValidateCertKeyPair(certData, keyData []byte, isSM2 bool) error {
	certs, err := v.ParseCertificates(certData, isSM2)
	if err != nil {
		return fmt.Errorf("解析证书失败: %w", err)
	}

	if len(certs) == 0 {
		return fmt.Errorf("证书链为空")
	}

	privKey, err := v.ParsePrivateKey(keyData, isSM2)
	if err != nil {
		return fmt.Errorf("解析密钥失败: %w", err)
	}

	leaf := certs[0]
	if err := v.validatePublicKey(leaf.PublicKey, privKey); err != nil {
		return err
	}

	return nil
}

// ParseCertificates 解析证书链
func (v *Validator) ParseCertificates(data []byte, isSM2 bool) ([]*x509.Certificate, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("无效的PEM格式")
	}

	if isSM2 {
		smCerts, err := smx509.ParseCertificates(block.Bytes)
		if err != nil {
			return nil, err
		}
		certs := make([]*x509.Certificate, len(smCerts))
		for i, smCert := range smCerts {
			certs[i] = smCert.ToX509()
		}
		return certs, nil
	}

	return x509.ParseCertificates(block.Bytes)
}

// ParsePrivateKey 解析私钥
func (v *Validator) ParsePrivateKey(data []byte, isSM2 bool) (crypto.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("无效的PEM格式")
	}

	if isSM2 {
		switch block.Type {
		case "PRIVATE KEY", "EC PRIVATE KEY":
			privKey, err := smx509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				privKey, err = smx509.ParseECPrivateKey(block.Bytes)
				if err != nil {
					return nil, fmt.Errorf("解析SM2私钥失败: %w", err)
				}
			}
			return privKey, nil
		default:
			privKey, err := smx509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("解析私钥失败: %w", err)
			}
			return privKey, nil
		}
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(block.Bytes)
	case "PRIVATE KEY":
		return x509.ParsePKCS8PrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("不支持的私钥类型: %s", block.Type)
	}
}

// validatePublicKey 验证公钥匹配
func (v *Validator) validatePublicKey(pub crypto.PublicKey, priv crypto.PrivateKey) error {
	switch pubKey := pub.(type) {
	case *rsa.PublicKey:
		privKey, ok := priv.(*rsa.PrivateKey)
		if !ok {
			return fmt.Errorf("证书公钥类型与私钥不匹配")
		}
		if pubKey.N.Cmp(privKey.N) != 0 || pubKey.E != privKey.E {
			return fmt.Errorf("RSA证书公钥与私钥不匹配")
		}
	case *ecdsa.PublicKey:
		privKey, ok := priv.(*ecdsa.PrivateKey)
		if !ok {
			return fmt.Errorf("证书公钥类型与私钥不匹配")
		}
		if !pubKey.Equal(privKey.Public()) {
			return fmt.Errorf("ECDSA证书公钥与私钥不匹配")
		}
	default:
		if pubKey == nil {
			return fmt.Errorf("证书公钥为空")
		}
	}
	return nil
}
