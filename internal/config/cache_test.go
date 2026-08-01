package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// setupCacheTest creates a temp config, populates it with the given
// content, and returns a cleanup func. It also invalidates any cache
// carried over from an earlier test.
func setupCacheTest(t *testing.T, content string) func() {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	SetConfigPathOverride(path)
	return func() { SetConfigPathOverride("") }
}

func TestReadCachedConfig_HitAvoidsDiskRead(t *testing.T) {
	cleanup := setupCacheTest(t, `{"name":"Alice","apiPort":9090}`)
	defer cleanup()

	first, err := readCachedConfig()
	if err != nil {
		t.Fatalf("first read: %v", err)
	}
	if first.Name != "Alice" || first.APIPort != 9090 {
		t.Fatalf("unexpected first read: %+v", first)
	}

	// Remove the underlying file. If the second call still hits disk,
	// it will fail. If the cache is doing its job, it succeeds.
	if err := os.Remove(GetConfigPath()); err != nil {
		t.Fatalf("remove config: %v", err)
	}

	second, err := readCachedConfig()
	if err != nil {
		t.Fatalf("second read should have hit cache: %v", err)
	}
	if second.Name != "Alice" || second.APIPort != 9090 {
		t.Fatalf("cache returned wrong data: %+v", second)
	}
}

func TestSaveConfig_InvalidatesCache(t *testing.T) {
	cleanup := setupCacheTest(t, `{"name":"Alice","apiPort":9090}`)
	defer cleanup()

	if _, err := readCachedConfig(); err != nil {
		t.Fatalf("warm cache: %v", err)
	}

	// Write via SaveConfig — the cache must reflect the new value.
	if err := SaveConfig(Config{Name: "Bob", APIPort: 7070}); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := readCachedConfig()
	if err != nil {
		t.Fatalf("read after save: %v", err)
	}
	if got.Name != "Bob" || got.APIPort != 7070 {
		t.Fatalf("cache stale after SaveConfig: %+v", got)
	}
}

func TestSetRuntimePort_InvalidatesCache(t *testing.T) {
	cleanup := setupCacheTest(t, `{"apiPort":9090}`)
	defer cleanup()

	if _, err := readCachedConfig(); err != nil {
		t.Fatalf("warm: %v", err)
	}

	SetRuntimePort(1234)
	defer SetRuntimePort(0)

	if GetAPIPort() != 1234 {
		t.Fatalf("runtime port not applied: %d", GetAPIPort())
	}
}

func TestSetConfigPathOverride_InvalidatesCache(t *testing.T) {
	// First config: Alice
	tmp1 := t.TempDir()
	p1 := filepath.Join(tmp1, "config.json")
	if err := os.WriteFile(p1, []byte(`{"name":"Alice"}`), 0600); err != nil {
		t.Fatal(err)
	}
	SetConfigPathOverride(p1)
	got1, err := readCachedConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got1.Name != "Alice" {
		t.Fatalf("first path: %+v", got1)
	}

	// Second config: Bob. Switching path must drop the Alice cache.
	tmp2 := t.TempDir()
	p2 := filepath.Join(tmp2, "config.json")
	if err := os.WriteFile(p2, []byte(`{"name":"Bob"}`), 0600); err != nil {
		t.Fatal(err)
	}
	SetConfigPathOverride(p2)
	defer SetConfigPathOverride("")

	got2, err := readCachedConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got2.Name != "Bob" {
		t.Fatalf("second path returned stale cache: %+v", got2)
	}
}

func TestReadCachedConfig_Concurrent(t *testing.T) {
	cleanup := setupCacheTest(t, `{"name":"Concurrent","apiPort":8080}`)
	defer cleanup()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				cfg, err := readCachedConfig()
				if err != nil {
					t.Errorf("read: %v", err)
					return
				}
				if cfg.Name != "Concurrent" {
					t.Errorf("wrong name: %q", cfg.Name)
					return
				}
			}
		}()
	}
	wg.Wait()
}
