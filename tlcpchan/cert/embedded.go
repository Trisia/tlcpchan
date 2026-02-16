package cert

import (
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"

	"github.com/emmansun/gmsm/smx509"
)

// EmbeddedCertLoader 嵌入式证书加载器，用于加载预制的根证书
type EmbeddedCertLoader struct {
	certDir     string
	trustedDir  string
	projectRoot string
}

// NewEmbeddedCertLoader 创建嵌入式证书加载器
// 参数:
//   - certDir: 证书目录路径，用于查找本地证书文件
//   - trustedDir: 受信任证书目录路径
//   - projectRoot: 项目根目录路径，用于查找预制证书
//
// 返回:
//   - *EmbeddedCertLoader: 嵌入式证书加载器实例
func NewEmbeddedCertLoader(certDir, trustedDir, projectRoot string) *EmbeddedCertLoader {
	return &EmbeddedCertLoader{
		certDir:     certDir,
		trustedDir:  trustedDir,
		projectRoot: projectRoot,
	}
}

// LoadDefaultCAs 加载默认的CA证书
// 返回:
//   - []string: CA证书文件路径列表
//   - error: 加载失败时返回错误
//
// 注意: 按以下顺序加载证书:
// 1. 用户指定的trustedDir目录中的证书
// 2. 项目根目录trustedcerts/中的预制证书
func (l *EmbeddedCertLoader) LoadDefaultCAs() ([]string, error) {
	var caPaths []string

	// 1. 首先尝试从用户指定的受信任证书目录加载
	if l.trustedDir != "" {
		patterns := []string{"*.crt", "*.cer", "*.pem"}
		for _, pattern := range patterns {
			matches, err := filepath.Glob(filepath.Join(l.trustedDir, pattern))
			if err != nil {
				continue
			}
			caPaths = append(caPaths, matches...)
		}
	}

	// 2. 如果用户目录没有证书，则尝试加载项目预制证书
	if len(caPaths) == 0 && l.projectRoot != "" {
		prebuiltDir := filepath.Join(l.projectRoot, "trustedcerts")
		patterns := []string{"*.crt", "*.cer", "*.pem"}
		for _, pattern := range patterns {
			matches, err := filepath.Glob(filepath.Join(prebuiltDir, pattern))
			if err != nil {
				continue
			}
			caPaths = append(caPaths, matches...)
		}
	}

	return caPaths, nil
}

// LoadSMCertPool 加载国密CA证书池
// 参数:
//   - caPaths: CA证书文件路径列表
//
// 返回:
//   - *smx509.CertPool: 国密证书池
//   - error: 读取或解析证书失败时返回错误
func (l *EmbeddedCertLoader) LoadSMCertPool(caPaths []string) (*smx509.CertPool, error) {
	pool := smx509.NewCertPool()

	for _, path := range caPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("读取CA证书失败 %s: %w", path, err)
		}

		if !pool.AppendCertsFromPEM(data) {
			return nil, fmt.Errorf("解析CA证书失败 %s", path)
		}
	}

	return pool, nil
}

// LoadX509CertPool 加载标准X509 CA证书池
// 参数:
//   - caPaths: CA证书文件路径列表
//   - includeSystem: 是否包含操作系统信任的根证书
//
// 返回:
//   - *x509.CertPool: X509证书池
//   - error: 读取或解析证书失败时返回错误
func (l *EmbeddedCertLoader) LoadX509CertPool(caPaths []string, includeSystem ...bool) (*x509.CertPool, error) {
	var pool *x509.CertPool

	// 如果需要，先加载系统根证书
	if len(includeSystem) > 0 && includeSystem[0] {
		pool, _ = x509.SystemCertPool()
	}
	if pool == nil {
		pool = x509.NewCertPool()
	}

	for _, path := range caPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("读取CA证书失败 %s: %w", path, err)
		}

		pool.AppendCertsFromPEM(data)
	}

	return pool, nil
}
