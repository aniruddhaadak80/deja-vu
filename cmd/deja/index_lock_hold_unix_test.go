//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/vshulcz/deja-vu/internal/index"
)

// holdTheIndexLock takes the lock the way another deja would, so the code under
// test takes the path it takes while a build is running. Its own file because
// the call below does not exist on windows, where the lock is taken through the
// share mode of the open itself.
func holdTheIndexLock(t *testing.T) {
	t.Helper()
	if os.Getenv("DEJA_INDEX_DIR") == "" {
		hermeticEnv(t)
	}
	path := index.DefaultDir() + ".lock"
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		t.Fatalf("could not hold the index lock: %v", err)
	}
	t.Cleanup(func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	})
}
