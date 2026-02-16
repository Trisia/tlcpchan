package keystore

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Store keystore 持久化存储
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
	return os.MkdirAll(s.baseDir, 0700)
}

func (s *Store) getRecordPath(name string) string {
	return filepath.Join(s.baseDir, name+".yaml")
}

func (s *Store) Save(record *StoreRecord) error {
	if err := s.ensureBaseDir(); err != nil {
		return err
	}

	data, err := yaml.Marshal(record)
	if err != nil {
		return fmt.Errorf("序列化记录失败: %w", err)
	}

	path := s.getRecordPath(record.Name)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("写入记录失败: %w", err)
	}

	return nil
}

func (s *Store) Load(name string) (*StoreRecord, error) {
	path := s.getRecordPath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取记录失败: %w", err)
	}

	var record StoreRecord
	if err := yaml.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("解析记录失败: %w", err)
	}

	return &record, nil
}

func (s *Store) LoadAll() ([]*StoreRecord, error) {
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

	var records []*StoreRecord
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
	path := s.getRecordPath(name)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除记录失败: %w", err)
	}
	return nil
}

func (s *Store) Exists(name string) bool {
	path := s.getRecordPath(name)
	_, err := os.Stat(path)
	return err == nil
}
