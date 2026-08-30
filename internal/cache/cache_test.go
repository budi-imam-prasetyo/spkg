package cache

import (
	"testing"
	"time"

	"github.com/ryoukaii/spkg/internal/detector"
)

func TestCacheOperations(t *testing.T) {
	Init(time.Minute, time.Minute)

	r := &detector.Result{
		Found:   true,
		Manager: detector.ManagerPacman,
		Package: "test-pkg",
		Version: "1.0.0",
	}

	Set("test-key", r)
	cached, found := Get("test-key")
	if !found {
		t.Fatal("expected to find cached item")
	}
	if cached.Package != "test-pkg" {
		t.Errorf("expected package 'test-pkg', got %q", cached.Package)
	}
}

func TestCacheMiss(t *testing.T) {
	Init(time.Minute, time.Minute)
	_, found := Get("nonexistent-key")
	if found {
		t.Error("expected cache miss")
	}
}

func TestCacheClear(t *testing.T) {
	Init(time.Minute, time.Minute)

	r := &detector.Result{Found: true}
	Set("key1", r)
	Set("key2", r)

	Clear()
	_, found := Get("key1")
	if found {
		t.Error("expected cache to be cleared")
	}
}

func TestCacheStats(t *testing.T) {
	Init(time.Minute, time.Minute)

	r := &detector.Result{Found: true}
	Set("key1", r)

	stats := Stats()
	if stats["items"] != 1 {
		t.Errorf("expected 1 item, got %d", stats["items"])
	}
}

func TestCacheSetWithTTL(t *testing.T) {
	Init(time.Minute, time.Minute)

	r := &detector.Result{Found: true, Package: "ttl-test"}
	SetWithTTL("ttl-key", r, time.Millisecond*50)

	cached, found := Get("ttl-key")
	if !found {
		t.Fatal("expected to find item before expiry")
	}
	if cached.Package != "ttl-test" {
		t.Errorf("expected 'ttl-test', got %q", cached.Package)
	}
}

func TestCacheItems(t *testing.T) {
	Init(time.Minute, time.Minute)

	r := &detector.Result{Found: true}
	Set("a", r)
	Set("b", r)

	items := Items()
	if len(items) < 2 {
		t.Errorf("expected at least 2 items, got %d", len(items))
	}
}

func BenchmarkCacheGet(b *testing.B) {
	Init(time.Minute, time.Minute)
	r := &detector.Result{Found: true, Package: "bench"}
	Set("bench-key", r)

	for i := 0; i < b.N; i++ {
		Get("bench-key")
	}
}

func BenchmarkCacheSet(b *testing.B) {
	Init(time.Minute, time.Minute)
	r := &detector.Result{Found: true, Package: "bench"}

	for i := 0; i < b.N; i++ {
		Set("bench-key", r)
	}
}
