package cache

import (
	"bytes"
	"encoding/gob"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/ryoukaii/spkg/internal/detector"
)

func init() {
	// Register the concrete type so gob can encode/decode *detector.Result.
	gob.Register(&detector.Result{})
}

var (
	saveMu sync.Mutex
	store  *gocache.Cache
	mu     sync.RWMutex
	ttl    time.Duration = 10 * time.Minute
	path   string
)

// Init creates the in-memory store and loads any persisted entries from disk.
func Init(defaultTTL, cleanupInterval time.Duration) {
	saveMu.Lock()
	defer saveMu.Unlock()
	mu.Lock()
	defer mu.Unlock()
	ttl = defaultTTL
	path = cachePath()
	store = gocache.NewFrom(defaultTTL, cleanupInterval, readDiskItems(path))
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
	full := filepath.Join(dir, "spkg", "cache-v2.gob")
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

// SetMany inserts a batch and persists it with a single serialization/write.
func SetMany(entries map[string]*detector.Result) {
	saveMu.Lock()
	defer saveMu.Unlock()
	mu.RLock()
	s, currentTTL := store, ttl
	mu.RUnlock()
	setEntries(s, entries, currentTTL)
}

func Set(key string, r *detector.Result) {
	SetMany(map[string]*detector.Result{key: r})
}

func SetWithTTL(key string, r *detector.Result, itemTTL time.Duration) {
	saveMu.Lock()
	defer saveMu.Unlock()
	mu.RLock()
	s := store
	mu.RUnlock()
	setEntries(s, map[string]*detector.Result{key: r}, itemTTL)
}

// Caller holds saveMu so mutations and persistence retain the same ordering.
func setEntries(s *gocache.Cache, entries map[string]*detector.Result, itemTTL time.Duration) {
	if s == nil || len(entries) == 0 {
		return
	}
	for k, r := range entries {
		s.Set(k, r, itemTTL)
	}
	updates := make(map[string]gocache.Item, len(entries))
	for k := range entries {
		if value, expiry, ok := s.GetWithExpiration(k); ok {
			item := gocache.Item{Object: value}
			if !expiry.IsZero() {
				item.Expiration = expiry.UnixNano()
			}
			updates[k] = item
		}
	}
	saveToDisk(updates, false)
}

func Clear() {
	saveMu.Lock()
	defer saveMu.Unlock()
	mu.RLock()
	s := store
	mu.RUnlock()
	if s == nil {
		return
	}
	s.Flush()
	saveToDisk(nil, true)
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

// saveToDisk merges only this operation's updates under a cross-process lock.
// Publishing a complete temporary file by rename keeps readers safe if a
// process exits while writing. Cache failures remain best-effort.
// Caller holds saveMu.
func saveToDisk(updates map[string]gocache.Item, clear bool) {
	mu.RLock()
	p := path
	mu.RUnlock()
	if p == "" {
		return
	}
	lock, err := os.OpenFile(p+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return
	}
	defer lock.Close()
	for {
		err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX)
		if err != syscall.EINTR {
			break
		}
	}
	if err != nil {
		return
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	items := make(map[string]gocache.Item)
	if !clear {
		items = readDiskItems(p)
	}
	for k, item := range updates {
		items[k] = item
	}
	data, err := encode(items)
	if err != nil {
		return
	}
	f, err := os.CreateTemp(filepath.Dir(p), ".cache-*.gob")
	if err != nil {
		return
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err != nil {
		f.Close()
		return
	}
	if err = f.Close(); err != nil {
		return
	}
	_ = os.Rename(f.Name(), p)
}

func encode(items map[string]gocache.Item) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(items); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Preserve absolute expiration timestamps rather than extending TTL on reload.
func readDiskItems(p string) map[string]gocache.Item {
	items := make(map[string]gocache.Item)
	data, err := os.ReadFile(p)
	if err != nil {
		return items
	}
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&items); err != nil {
		return make(map[string]gocache.Item)
	}
	if items == nil {
		return make(map[string]gocache.Item)
	}
	now := time.Now().UnixNano()
	for k, item := range items {
		if item.Expiration > 0 && item.Expiration <= now {
			delete(items, k)
		}
	}
	return items
}
