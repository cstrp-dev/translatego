package ratelimit

import (
	"sync"
	"time"
)

type Manager struct {
	limiters map[string]*Limiter
	mu       sync.RWMutex
}

type Limiter struct {
	Requests    int
	LastReset   time.Time
	MaxRequests int
	Window      time.Duration
}

func NewManager() *Manager {
	return &Manager{
		limiters: make(map[string]*Limiter),
	}
}

func (m *Manager) GetLimiter(serviceName string, maxRequests int, window time.Duration) *Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	if limiter, exists := m.limiters[serviceName]; exists {
		return limiter
	}

	limiter := &Limiter{
		Requests:    0,
		LastReset:   time.Now(),
		MaxRequests: maxRequests,
		Window:      window,
	}
	m.limiters[serviceName] = limiter
	return limiter
}

func (m *Manager) Allow(serviceName string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	limiter, exists := m.limiters[serviceName]
	if !exists {
		limiter = &Limiter{
			Requests:    0,
			LastReset:   time.Now(),
			MaxRequests: 10,
			Window:      time.Minute,
		}
		m.limiters[serviceName] = limiter
	}

	return limiter.Allow()
}

func (m *Manager) RecordRequest(serviceName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if limiter, exists := m.limiters[serviceName]; exists {
		limiter.RecordRequest()
	}
}

func (l *Limiter) Allow() bool {
	now := time.Now()
	if now.Sub(l.LastReset) > l.Window {
		l.Requests = 0
		l.LastReset = now
	}
	return l.Requests < l.MaxRequests
}

func (l *Limiter) RecordRequest() {
	l.Requests++
}
