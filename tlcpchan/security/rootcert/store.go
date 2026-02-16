package rootcert

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Store 根证书持久化存储
type Store struct {
	baseDir string
}

func NewStore(baseDir string) *Store {
	return &Store{baseDir: baseDir}
}

func (s *Store) ensureBaseDir() error {
	if s.baseDir == "" {
		return nil
	}
	if err := os.MkdirAll(s.baseDir, 0700); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(s.baseDir, "certs"), 0700); err != nil {
		return err
	}
	return nil
}

func (s *Store) getRecordPath(name string) string {
	return filepath.Join(s.baseDir, name+".yaml")
}

func (s *Store) getCertPath(name string) string {
	return filepath.Join(s.baseDir, "certs", name+".pem")
}

type storeRecord struct {
	Name    string    `yaml:"name"`
	AddedAt time.Time `yaml:"addedAt"`
}

func (s *Store) Save(name string, certData []byte, addedAt time.Time) error {
	if err := s.ensureBaseDir(); err != nil {
		return err
	}

	record := &storeRecord{
		Name:    name,
		AddedAt: addedAt,
	}

	recordData, err := yaml.Marshal(record)
	if err != nil {
		return fmt.Errorf("序列化记录失败: %w", err)
	}

	if err := os.WriteFile(s.getRecordPath(name), recordData, 0600); err != nil {
		return fmt.Errorf("写入记录失败: %w", err)
	}

	if err := os.WriteFile(s.getCertPath(name), certData, 0600); err != nil {
		return fmt.Errorf("写入证书失败: %w", err)
	}

	return nil
}

func (s *Store) Load(name string) (*storeRecord, error) {
	recordPath := s.getRecordPath(name)
	data, err := os.ReadFile(recordPath)
	if err != nil {
		return nil, fmt.Errorf("读取记录失败: %w", err)
	}

	var record storeRecord
	if err := yaml.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("解析记录失败: %w", err)
	}

	return &record, nil
}

func (s *Store) LoadAll() ([]*storeRecord, error) {
	if s.baseDir == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}

	var records []*storeRecord
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) < 6 || name[len(name)-5:] != ".yaml" {
			continue
		}

		recordName := name[:len(name)-5]
		record, err := s.Load(recordName)
		if err != nil {
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

func (s *Store) Delete(name string) error {
	if err := os.Remove(s.getRecordPath(name)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除记录失败: %w", err)
	}
	if err := os.Remove(s.getCertPath(name)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除证书失败: %w", err)
	}
	return nil
}
