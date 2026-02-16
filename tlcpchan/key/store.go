package key

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Store 密钥存储操作
type Store struct {
	baseDir string
}

// NewStore 创建新的存储操作实例
func NewStore(baseDir string) *Store {
	return &Store{
		baseDir: baseDir,
	}
}

// EnsureBaseDir 确保基础目录存在
func (s *Store) EnsureBaseDir() error {
	return os.MkdirAll(s.baseDir, 0755)
}

// getKeyDir 获取密钥目录路径
func (s *Store) getKeyDir(name string) string {
	return filepath.Join(s.baseDir, name)
}

// Exists 检查密钥是否存在
func (s *Store) Exists(name string) bool {
	infoPath := filepath.Join(s.getKeyDir(name), FileNameInfo)
	_, err := os.Stat(infoPath)
	return err == nil
}

// Save 保存密钥信息和文件
func (s *Store) Save(ks *KeyStore, signCertData, signKeyData, encCertData, encKeyData []byte) error {
	keyDir := s.getKeyDir(ks.Name)
	if err := os.MkdirAll(keyDir, 0755); err != nil {
		return fmt.Errorf("创建密钥目录失败: %w", err)
	}

	infoPath := filepath.Join(keyDir, FileNameInfo)
	data, err := yaml.Marshal(ks)
	if err != nil {
		return fmt.Errorf("序列化密钥信息失败: %w", err)
	}
	if err := os.WriteFile(infoPath, data, 0644); err != nil {
		return fmt.Errorf("保存密钥信息失败: %w", err)
	}

	if len(signCertData) > 0 {
		if err := os.WriteFile(filepath.Join(keyDir, FileNameSignCert), signCertData, 0644); err != nil {
			return fmt.Errorf("保存签名证书失败: %w", err)
		}
	}

	if len(signKeyData) > 0 {
		if err := os.WriteFile(filepath.Join(keyDir, FileNameSignKey), signKeyData, 0600); err != nil {
			return fmt.Errorf("保存签名密钥失败: %w", err)
		}
	}

	if len(encCertData) > 0 {
		if err := os.WriteFile(filepath.Join(keyDir, FileNameEncCert), encCertData, 0644); err != nil {
			return fmt.Errorf("保存加密证书失败: %w", err)
		}
	}

	if len(encKeyData) > 0 {
		if err := os.WriteFile(filepath.Join(keyDir, FileNameEncKey), encKeyData, 0600); err != nil {
			return fmt.Errorf("保存加密密钥失败: %w", err)
		}
	}

	return nil
}

// UpdateCertificates 仅更新证书
func (s *Store) UpdateCertificates(name string, signCertData, encCertData []byte) error {
	keyDir := s.getKeyDir(name)
	if _, err := os.Stat(keyDir); os.IsNotExist(err) {
		return fmt.Errorf("密钥不存在")
	}

	if len(signCertData) > 0 {
		if err := os.WriteFile(filepath.Join(keyDir, FileNameSignCert), signCertData, 0644); err != nil {
			return fmt.Errorf("更新签名证书失败: %w", err)
		}
	}

	if len(encCertData) > 0 {
		if err := os.WriteFile(filepath.Join(keyDir, FileNameEncCert), encCertData, 0644); err != nil {
			return fmt.Errorf("更新加密证书失败: %w", err)
		}
	}

	ks, err := s.Load(name)
	if err != nil {
		return fmt.Errorf("加载密钥信息失败: %w", err)
	}
	ks.UpdatedAt = ks.UpdatedAt.Local()

	infoPath := filepath.Join(keyDir, FileNameInfo)
	data, err := yaml.Marshal(ks)
	if err != nil {
		return fmt.Errorf("序列化密钥信息失败: %w", err)
	}
	if err := os.WriteFile(infoPath, data, 0644); err != nil {
		return fmt.Errorf("保存密钥信息失败: %w", err)
	}

	return nil
}

// Load 加载密钥信息
func (s *Store) Load(name string) (*KeyStore, error) {
	infoPath := filepath.Join(s.getKeyDir(name), FileNameInfo)
	data, err := os.ReadFile(infoPath)
	if err != nil {
		return nil, fmt.Errorf("读取密钥信息失败: %w", err)
	}

	var ks KeyStore
	if err := yaml.Unmarshal(data, &ks); err != nil {
		return nil, fmt.Errorf("解析密钥信息失败: %w", err)
	}

	return &ks, nil
}

// Delete 删除密钥
func (s *Store) Delete(name string) error {
	keyDir := s.getKeyDir(name)
	if _, err := os.Stat(keyDir); os.IsNotExist(err) {
		return fmt.Errorf("密钥不存在")
	}
	return os.RemoveAll(keyDir)
}

// List 列出所有密钥
func (s *Store) List() ([]*KeyStore, error) {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*KeyStore{}, nil
		}
		return nil, fmt.Errorf("读取密钥目录失败: %w", err)
	}

	var result []*KeyStore
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		ks, err := s.Load(name)
		if err != nil {
			continue
		}
		result = append(result, ks)
	}

	return result, nil
}

// GetFilePath 获取文件绝对路径
func (s *Store) GetFilePath(name, fileName string) string {
	return filepath.Join(s.getKeyDir(name), fileName)
}

// HasFile 检查文件是否存在
func (s *Store) HasFile(name, fileName string) bool {
	path := s.GetFilePath(name, fileName)
	_, err := os.Stat(path)
	return err == nil
}
