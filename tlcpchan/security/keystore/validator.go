package keystore

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"fmt"

	"github.com/Trisia/tlcpchan/security/der"
	"github.com/emmansun/gmsm/smx509"
)

// VerifyCertificateKeyPair 验证证书和私钥是否匹配
//
// 参数：
//   - certData: 证书数据（PEM、DER、HEX 或 Base64 格式）
//   - keyData: 私钥数据（PEM、DER、HEX 或 Base64 格式）
//   - isTLCP: 是否为 TLCP 类型（true 使用 SM2，false 使用 RSA/ECDSA）
//
// 返回：
//   - error: 验证失败返回错误信息
//
// 注意事项：
//   - 自动调用 Any2DER 转换数据格式
//   - TLCP 类型使用 SM2 算法
//   - TLS 类型支持 RSA 和 ECDSA 算法
func VerifyCertificateKeyPair(certData, keyData []byte, isTLCP bool) error {
	certDER, err := der.Any2DER(certData)
	if err != nil {
		return fmt.Errorf("解析证书失败: %w", err)
	}

	keyDER, err := der.Any2DER(keyData)
	if err != nil {
		return fmt.Errorf("解析私钥失败: %w", err)
	}

	var certPub crypto.PublicKey
	var privKey crypto.PrivateKey

	if isTLCP {
		smCerts, err := smx509.ParseCertificates(certDER)
		if err != nil || len(smCerts) == 0 {
			return fmt.Errorf("解析证书失败: %w", err)
		}
		certPub = smCerts[0].ToX509().PublicKey

		privKey, err = smx509.ParsePKCS8PrivateKey(keyDER)
		if err != nil {
			privKey, err = smx509.ParseECPrivateKey(keyDER)
			if err != nil {
				return fmt.Errorf("解析私钥失败: %w", err)
			}
		}
	} else {
		x509Cert, err := x509.ParseCertificate(certDER)
		if err != nil {
			return fmt.Errorf("解析证书失败: %w", err)
		}
		certPub = x509Cert.PublicKey

		privKey, err = x509.ParsePKCS8PrivateKey(keyDER)
		if err != nil {
			privKey, err = x509.ParseECPrivateKey(keyDER)
			if err != nil {
				privKey, err = x509.ParsePKCS1PrivateKey(keyDER)
				if err != nil {
					return fmt.Errorf("解析私钥失败: %w", err)
				}
			}
		}
	}

	privPubKey := privKey.(interface{ Public() crypto.PublicKey }).Public()

	switch certPub := certPub.(type) {
	case *ecdsa.PublicKey:
		if privPub, ok := privPubKey.(*ecdsa.PublicKey); ok {
			if certPub.X.Cmp(privPub.X) == 0 && certPub.Y.Cmp(privPub.Y) == 0 {
				return nil
			}
		}
		return fmt.Errorf("ECDSA 公钥不匹配")
	case *rsa.PublicKey:
		if privPub, ok := privPubKey.(*rsa.PublicKey); ok {
			if certPub.N.Cmp(privPub.N) == 0 && certPub.E == privPub.E {
				return nil
			}
		}
		return fmt.Errorf("RSA 公钥不匹配")
	default:
		return fmt.Errorf("不支持的公钥类型")
	}
}
