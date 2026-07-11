package cache

import (
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/ryoukaii/spkg/internal/detector"
)

var (
	store *gocache.Cache
	mu    sync.RWMutex
	ttl   time.Duration = 10 * time.Minute
)

func Init(defaultTTL, cleanupInterval time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	ttl = defaultTTL
	store = gocache.New(defaultTTL, cleanupInterval)
}

func SetTTL(newTTL time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	ttl = newTTL
}

func Get(key string) (*detector.Result, bool) {
	mu.RLock()
	s := store
	mu.RUnlock()
	if s == nil {
		return nil, false
	}
	v, found := s.Get(key)
	if !found {
		return nil, false
	}
	r, ok := v.(*detector.Result)
	return r, ok
}

func Set(key string, r *detector.Result) {
	mu.RLock()
	s := store
	currentTTL := ttl
	mu.RUnlock()
	if s == nil {
		return
	}
	s.Set(key, r, currentTTL)
}

func SetWithTTL(key string, r *detector.Result, itemTTL time.Duration) {
	mu.RLock()
	s := store
	mu.RUnlock()
	if s == nil {
		return
	}
	s.Set(key, r, itemTTL)
}

func Clear() {
	mu.RLock()
	s := store
	mu.RUnlock()
	if s == nil {
		return
	}
	s.Flush()
}

func Stats() map[string]int {
	mu.RLock()
	s := store
	currentTTL := ttl
	mu.RUnlock()
	if s == nil {
		return map[string]int{"items": 0}
	}
	return map[string]int{
		"items":      s.ItemCount(),
		"ttl_secs":   int(currentTTL.Seconds()),
	}
}

func Items() map[string]gocache.Item {
	mu.RLock()
	s := store
	mu.RUnlock()
	if s == nil {
		return nil
	}
	return s.Items()
}
