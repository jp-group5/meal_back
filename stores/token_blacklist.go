package stores

import (
	"sync"
	"time"
)

// TokenBlacklistStore 管理已失效 token 的 jti（内存版）。
// 说明：生产环境建议替换为 Redis，以支持多实例部署。
type TokenBlacklistStore struct {
	mu    sync.RWMutex
	items map[string]time.Time
}

// NewTokenBlacklistStore 创建黑名单存储。
func NewTokenBlacklistStore() *TokenBlacklistStore {
	return &TokenBlacklistStore{
		items: make(map[string]time.Time),
	}
}

// Add 将 token jti 加入黑名单，直到过期时间。
func (s *TokenBlacklistStore) Add(jti string, expiresAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[jti] = expiresAt
}

// IsBlacklisted 判断 token jti 是否已失效。
func (s *TokenBlacklistStore) IsBlacklisted(jti string) bool {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	expiresAt, ok := s.items[jti]
	if !ok {
		return false
	}

	// 已过期的黑名单记录及时清理。
	if now.After(expiresAt) {
		delete(s.items, jti)
		return false
	}
	return true
}
