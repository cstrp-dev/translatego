package cache

import (
	"fmt"
	"sync"
)

type Manager struct {
	cache map[string]map[string]string
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		cache: make(map[string]map[string]string),
	}
}

func (m *Manager) GetCacheKey(text, source, target string) string {
	return fmt.Sprintf("%s|%s|%s", text, source, target)
}

func (m *Manager) Get(service, text, source, target string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if serviceCache, exists := m.cache[service]; exists {
		cacheKey := m.GetCacheKey(text, source, target)
		if translation, cached := serviceCache[cacheKey]; cached {
			return translation, true
		}
	}
	return "", false
}

func (m *Manager) Set(service, text, source, target, translation string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.cache[service]; !exists {
		m.cache[service] = make(map[string]string)
	}
	cacheKey := m.GetCacheKey(text, source, target)
	m.cache[service][cacheKey] = translation
}

func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache = make(map[string]map[string]string)
}

func (m *Manager) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	total := 0
	for _, serviceCache := range m.cache {
		total += len(serviceCache)
	}
	return total
}
