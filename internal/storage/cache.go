package storage

import (
	"sync"
	"time"
)

// CacheItem 缓存项
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
}

// MemoryCache 内存缓存实现
type MemoryCache struct {
	items map[string]*CacheItem
	mutex sync.RWMutex
}

// NewMemoryCache 创建新的内存缓存实例
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]*CacheItem),
	}

	// 启动清理过期项的 goroutine
	go cache.cleanup()

	return cache
}

// Set 设置缓存项
func (mc *MemoryCache) Set(key string, value interface{}, duration time.Duration) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	expiresAt := time.Now().Add(duration)
	mc.items[key] = &CacheItem{
		Value:     value,
		ExpiresAt: expiresAt,
	}

	return nil
}

// Get 获取缓存项
func (mc *MemoryCache) Get(key string) (interface{}, bool) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	item, exists := mc.items[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		// 延迟删除过期项
		go func() {
			mc.mutex.Lock()
			delete(mc.items, key)
			mc.mutex.Unlock()
		}()
		return nil, false
	}

	return item.Value, true
}

// Delete 删除缓存项
func (mc *MemoryCache) Delete(key string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	delete(mc.items, key)
}

// Clear 清空所有缓存项
func (mc *MemoryCache) Clear() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.items = make(map[string]*CacheItem)
}

// Size 获取缓存项数量
func (mc *MemoryCache) Size() int {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	return len(mc.items)
}

// Keys 获取所有缓存键
func (mc *MemoryCache) Keys() []string {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	keys := make([]string, 0, len(mc.items))
	for key := range mc.items {
		keys = append(keys, key)
	}

	return keys
}

// cleanup 定期清理过期的缓存项
func (mc *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		mc.cleanupExpired()
	}
}

// cleanupExpired 清理过期的缓存项
func (mc *MemoryCache) cleanupExpired() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	now := time.Now()
	for key, item := range mc.items {
		if now.After(item.ExpiresAt) {
			delete(mc.items, key)
		}
	}
}

// GetWithTTL 获取缓存项和剩余过期时间
func (mc *MemoryCache) GetWithTTL(key string) (interface{}, time.Duration, bool) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	item, exists := mc.items[key]
	if !exists {
		return nil, 0, false
	}

	now := time.Now()
	if now.After(item.ExpiresAt) {
		// 延迟删除过期项
		go func() {
			mc.mutex.Lock()
			delete(mc.items, key)
			mc.mutex.Unlock()
		}()
		return nil, 0, false
	}

	ttl := item.ExpiresAt.Sub(now)
	return item.Value, ttl, true
}
