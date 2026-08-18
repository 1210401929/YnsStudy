package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"nys-go-api/internal/config"
)

var ErrNotFound = errors.New("缓存键不存在")

type Store interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Increment(ctx context.Context, key string, ttl time.Duration) (int64, error)
	Close() error
}

func NewStore(ctx context.Context, cfg config.RedisConfig) (Store, error) {
	if !cfg.Enabled {
		return NewMemoryStore(), nil
	}
	client := redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:    cfg.Password,
		DB:          cfg.Database,
		DialTimeout: time.Duration(cfg.DialTimeoutSeconds) * time.Second,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("连接 Redis: %w", err)
	}
	return &redisStore{client: client}, nil
}

type redisStore struct {
	client *redis.Client
}

var incrementWithExpiration = redis.NewScript(`
local count = redis.call("INCR", KEYS[1])
if count == 1 then
    redis.call("EXPIRE", KEYS[1], ARGV[1])
end
return count
`)

func (s *redisStore) Get(ctx context.Context, key string) (string, error) {
	value, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	}
	return value, err
}

func (s *redisStore) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

func (s *redisStore) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

func (s *redisStore) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	if ttl <= 0 {
		return s.client.Incr(ctx, key).Result()
	}

	// Lua 保证自增和首次设置过期时间的原子性，同时兼容旧版本 Redis。
	expirationSeconds := int64((ttl + time.Second - 1) / time.Second)
	return incrementWithExpiration.Run(ctx, s.client, []string{key}, expirationSeconds).Int64()
}

func (s *redisStore) Close() error {
	return s.client.Close()
}

type memoryEntry struct {
	value     string
	expiresAt time.Time
}

type MemoryStore struct {
	mu      sync.Mutex
	entries map[string]memoryEntry
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{entries: make(map[string]memoryEntry)}
}

func (s *MemoryStore) Get(_ context.Context, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[key]
	if !ok || (!entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt)) {
		delete(s.entries, key)
		return "", ErrNotFound
	}
	return entry.value, nil
}

func (s *MemoryStore) Set(_ context.Context, key, value string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := memoryEntry{value: value}
	if ttl > 0 {
		entry.expiresAt = time.Now().Add(ttl)
	}
	s.entries[key] = entry
	return nil
}

func (s *MemoryStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
	return nil
}

func (s *MemoryStore) Increment(_ context.Context, key string, ttl time.Duration) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	entry, ok := s.entries[key]
	if !ok || (!entry.expiresAt.IsZero() && now.After(entry.expiresAt)) {
		entry = memoryEntry{value: "0", expiresAt: now.Add(ttl)}
	}
	var count int64
	_, _ = fmt.Sscan(entry.value, &count)
	count++
	entry.value = fmt.Sprint(count)
	s.entries[key] = entry
	return count, nil
}

func (s *MemoryStore) Close() error { return nil }
