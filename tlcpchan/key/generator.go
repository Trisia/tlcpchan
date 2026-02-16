package key

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/smx509"
)

// Generator 密钥生成器
type Generator struct{}

// NewGenerator 创建密钥生成器
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateSM2Key 生成SM2密钥对
func (g *Generator) GenerateSM2Key() ([]byte, []byte, error) {
	privKey, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("生成SM2密钥失败: %w", err)
	}

	keyBytes, err := smx509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, nil, fmt.Errorf("序列化SM2私钥失败: %w", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	})

	pubKeyBytes, err := smx509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("序列化SM2公钥失败: %w", err)
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return keyPEM, pubPEM, nil
}

// GenerateRSAKey 生成RSA密钥对
func (g *Generator) GenerateRSAKey(bits int) ([]byte, []byte, error) {
	if bits <= 0 {
		bits = 2048
	}

	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("生成RSA密钥失败: %w", err)
	}

	keyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	})

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("序列化RSA公钥失败: %w", err)
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return keyPEM, pubPEM, nil
}

// GenerateECDSAKey 生成ECDSA密钥对
func (g *Generator) GenerateECDSAKey(curve elliptic.Curve) ([]byte, []byte, error) {
	if curve == nil {
		curve = elliptic.P256()
	}

	privKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("生成ECDSA密钥失败: %w", err)
	}

	keyBytes, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return nil, nil, fmt.Errorf("序列化ECDSA私钥失败: %w", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	})

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("序列化ECDSA公钥失败: %w", err)
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return keyPEM, pubPEM, nil
}

// GenerateKeyByParams 根据参数生成密钥
func (g *Generator) GenerateKeyByParams(params KeyParams) ([]byte, []byte, error) {
	switch params.Algorithm {
	case "SM2", "sm2":
		return g.GenerateSM2Key()
	case "RSA", "rsa":
		bits := params.Length
		if bits <= 0 {
			bits = 2048
		}
		return g.GenerateRSAKey(bits)
	case "ECDSA", "ecdsa":
		var curve elliptic.Curve
		switch params.Length {
		case 256:
			curve = elliptic.P256()
		case 384:
			curve = elliptic.P384()
		case 521:
			curve = elliptic.P521()
		default:
			curve = elliptic.P256()
		}
		return g.GenerateECDSAKey(curve)
	default:
		return nil, nil, fmt.Errorf("不支持的算法: %s", params.Algorithm)
	}
}
