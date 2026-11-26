package cache

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/ryoukaii/spkg/internal/detector"
)

func TestSetManyPersistsBatchAndExpiration(t *testing.T) {
	initTestCache(t)
	SetMany(map[string]*detector.Result{
		"a": {Found: true, Package: "first"},
		"b": {Found: true, Package: "second"},
	})
	SetWithTTL("permanent", &detector.Result{Found: true}, gocache.NoExpiration)
	before := Items()["a"].Expiration
	Init(time.Minute, 0)
	for _, key := range []string{"a", "b", "permanent"} {
		if _, ok := Get(key); !ok {
			t.Fatalf("%s missing after reload", key)
		}
	}
	if expiry := Items()["a"].Expiration; expiry < before || expiry-before > int64(time.Millisecond) {
		t.Fatalf("reload changed expiration: %d -> %d", before, expiry)
	}
	if expiry := Items()["permanent"].Expiration; expiry != 0 {
		t.Fatalf("non-expiring item now expires at %d", expiry)
	}
	Clear()
	Init(time.Minute, 0)
	if len(Items()) != 0 {
		t.Fatal("clear did not persist")
	}
}

func TestConcurrentPersistenceKeepsFileReadable(t *testing.T) {
	initTestCache(t)
	Set("seed", &detector.Result{Found: true})
	cacheFile := path
	stop := make(chan struct{})
	readDone := make(chan error, 1)
	go func() {
		for {
			select {
			case <-stop:
				readDone <- nil
				return
			default:
			}
			data, err := os.ReadFile(cacheFile)
			if err == nil {
				var items map[string]gocache.Item
				err = gob.NewDecoder(bytes.NewReader(data)).Decode(&items)
			}
			if err != nil {
				readDone <- err
				return
			}
		}
	}()
	var wg sync.WaitGroup
	for i := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Set(fmt.Sprint(i), &detector.Result{Found: true})
		}()
	}
	wg.Wait()
	close(stop)
	if err := <-readDone; err != nil {
		t.Fatalf("reader observed an incomplete cache: %v", err)
	}
	Init(time.Minute, 0)
	if count := len(Items()); count != 17 {
		t.Fatalf("persisted %d items, want 17", count)
	}
}

// Both writers initialize before either writes, reproducing the stale snapshot
// that would lose one process's additions without a locked merge.
func TestPersistenceMergesIndependentProcesses(t *testing.T) {
	initTestCache(t)
	type writer struct {
		cmd    *exec.Cmd
		start  func()
		stderr *bytes.Buffer
	}
	var writers []writer
	for _, key := range []string{"process-a", "process-b"} {
		cmd := exec.Command(os.Args[0], "-test.run=^TestCacheWriterProcess$")
		cmd.Env = append(os.Environ(), "SPKG_TEST_CACHE_KEY="+key)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = cmd.Process.Kill() })
		ready, err := bufio.NewReader(stdout).ReadString('\n')
		if err != nil || ready != "ready\n" {
			t.Fatalf("writer startup: %q %v %s", ready, err, stderr.String())
		}
		writers = append(writers, writer{cmd, func() { _, _ = stdin.Write([]byte("start\n")); _ = stdin.Close() }, &stderr})
	}
	for _, w := range writers {
		w.start()
	}
	for _, w := range writers {
		if err := w.cmd.Wait(); err != nil {
			t.Fatalf("writer failed: %v %s", err, w.stderr.String())
		}
	}
	Init(time.Minute, 0)
	for _, key := range []string{"process-a", "process-b"} {
		if _, ok := Get(key); !ok {
			t.Fatalf("lost concurrent writer's entry %s", key)
		}
	}
}

func TestCacheWriterProcess(t *testing.T) {
	key := os.Getenv("SPKG_TEST_CACHE_KEY")
	if key == "" {
		return
	}
	Init(time.Minute, 0)
	fmt.Println("ready")
	if _, err := bufio.NewReader(os.Stdin).ReadString('\n'); err != nil {
		t.Fatal(err)
	}
	SetMany(map[string]*detector.Result{key: {Found: true, Package: key}})
}
