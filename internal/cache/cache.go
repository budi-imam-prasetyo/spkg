package cache

import (
	"bytes"
	"encoding/gob"
	"os"
	"path/filepath"
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/ryoukaii/spkg/internal/detector"
)

func init() {
	// Register the concrete type so gob can encode/decode *detector.Result.
	gob.Register(&detector.Result{})
}

var (
	store *gocache.Cache
	mu    sync.RWMutex
	ttl   time.Duration = 10 * time.Minute
	path  string
)

// Init creates the in-memory store and loads any persisted entries from disk.
func Init(defaultTTL, cleanupInterval time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	ttl = defaultTTL
	store = gocache.New(defaultTTL, cleanupInterval)
	path = cachePath()
	if path != "" {
		loadFromDisk(path, store)
	}
}

func SetTTL(newTTL time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	ttl = newTTL
}

func cachePath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return ""
		}
		dir = filepath.Join(home, ".cache")
	}
	full := filepath.Join(dir, "spkg", "cache.gob")
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return ""
	}
	return full
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
	saveToDisk()
}

func SetWithTTL(key string, r *detector.Result, itemTTL time.Duration) {
	mu.RLock()
	s := store
	mu.RUnlock()
	if s == nil {
		return
	}
	s.Set(key, r, itemTTL)
	saveToDisk()
}

func Clear() {
	mu.RLock()
	s := store
	mu.RUnlock()
	if s == nil {
		return
	}
	s.Flush()
	saveToDisk()
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
		"items":    s.ItemCount(),
		"ttl_secs": int(currentTTL.Seconds()),
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

// saveToDisk persists the whole cache to a gob file so it survives across `spkg`
// invocations. Errors are intentionally ignored — caching is best-effort.
func saveToDisk() {
	if path == "" {
		return
	}
	mu.RLock()
	s := store
	mu.RUnlock()
	if s == nil {
		return
	}
	data, err := encode(s.Items())
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o600)
}

func encode(items map[string]gocache.Item) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(items); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func loadFromDisk(p string, s *gocache.Cache) {
	data, err := os.ReadFile(p)
	if err != nil {
		return
	}
	var items map[string]gocache.Item
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&items); err != nil {
		return
	}
	now := time.Now()
	for k, item := range items {
		if item.Expiration > 0 && item.Expiration < now.UnixNano() {
			continue
		}
		s.Set(k, item.Object, time.Duration(item.Expiration-now.UnixNano()))
	}
}
