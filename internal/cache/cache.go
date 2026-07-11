package cache

import (
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/ryoukaii/spkg/internal/detector"
)

var (
	store *gocache.Cache
	once  sync.Once
)

func Init(defaultTTL, cleanupInterval time.Duration) {
	once.Do(func() {
		store = gocache.New(defaultTTL, cleanupInterval)
	})
}

func Get(key string) (*detector.Result, bool) {
	if store == nil {
		return nil, false
	}
	v, found := store.Get(key)
	if !found {
		return nil, false
	}
	r, ok := v.(*detector.Result)
	return r, ok
}

func Set(key string, r *detector.Result) {
	if store == nil {
		return
	}
	store.Set(key, r, gocache.DefaultExpiration)
}

func SetWithTTL(key string, r *detector.Result, ttl time.Duration) {
	if store == nil {
		return
	}
	store.Set(key, r, ttl)
}

func Clear() {
	if store == nil {
		return
	}
	store.Flush()
}

func Stats() map[string]int {
	if store == nil {
		return map[string]int{"items": 0}
	}
	return map[string]int{
		"items": store.ItemCount(),
	}
}

func Items() map[string]gocache.Item {
	if store == nil {
		return nil
	}
	return store.Items()
}
