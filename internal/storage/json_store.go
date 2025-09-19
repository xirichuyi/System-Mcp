package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// JSONStorage JSON 文件存储实现
type JSONStorage struct {
	dataDir string
	mutex   sync.RWMutex
}

// NewJSONStorage 创建新的 JSON 存储实例
func NewJSONStorage(dataDir string) (*JSONStorage, error) {
	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}

	return &JSONStorage{
		dataDir: dataDir,
	}, nil
}

// Save 保存数据到 JSON 文件
func (js *JSONStorage) Save(key string, data interface{}) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	filePath := filepath.Join(js.dataDir, key+".json")

	// 序列化数据
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// Load 从 JSON 文件加载数据
func (js *JSONStorage) Load(key string, data interface{}) error {
	js.mutex.RLock()
	defer js.mutex.RUnlock()

	filePath := filepath.Join(js.dataDir, key+".json")

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// 读取文件
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// 反序列化数据
	if err := json.Unmarshal(jsonData, data); err != nil {
		return fmt.Errorf("failed to unmarshal data: %v", err)
	}

	return nil
}

// Delete 删除 JSON 文件
func (js *JSONStorage) Delete(key string) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	filePath := filepath.Join(js.dataDir, key+".json")

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}

// Exists 检查文件是否存在
func (js *JSONStorage) Exists(key string) bool {
	js.mutex.RLock()
	defer js.mutex.RUnlock()

	filePath := filepath.Join(js.dataDir, key+".json")
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// ListKeys 列出所有存储的键
func (js *JSONStorage) ListKeys() ([]string, error) {
	js.mutex.RLock()
	defer js.mutex.RUnlock()

	files, err := os.ReadDir(js.dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	var keys []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			key := file.Name()[:len(file.Name())-5] // 去掉 .json 扩展名
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// GetDataDir 获取数据目录路径
func (js *JSONStorage) GetDataDir() string {
	return js.dataDir
}
