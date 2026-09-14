package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"

	"github.com/vshulcz/deja-vu/internal/index"
	"github.com/vshulcz/deja-vu/internal/query"
)

// A search that lands while another process holds the index lock is told to
// serve what is on disk. On a first build there is nothing on disk, and the run
// said two things that cannot both be true:
//
//	deja: answering from the index as it was — refreshing in the background
//	deja: the index is being rebuilt right now — run this again in a moment
//
// That is a new machine's first question, which is the worst run to contradict
// anything on (#3574).
func TestTheFirstBuildDoesNotClaimToAnswerFromAnIndex(t *testing.T) {
	holdTheIndexLock(t)
	var said bytes.Buffer
	if err := ensureForCLISearch(index.DefaultDir(), query.Options{Query: "retry"}, false, &said); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if strings.Contains(said.String(), "as it was") {
		t.Fatalf("an empty store claimed to be answering from an index:\n%s", said.String())
	}
}

// With a store that answers, the line is right and stays: that is the case it
// was written for — an index serving while the refresh runs behind it.
func TestAStoreThatAnswersStillSaysItIsServingWhatItHas(t *testing.T) {
	tmp := hermeticEnv(t)
	writeClaudeFixture(t, filepath.Join(tmp, "claude", "app", "one.jsonl"), "s1", []string{
		`{"type":"user","sessionId":"s1","timestamp":"2026-09-01T10:00:00Z","message":{"role":"user","content":"the exporter drops every third retry"}}`,
	})
	if _, err := captureRunStderr(t, "index"); err != nil {
		t.Fatal(err)
	}
	// A transcript that has appeared since, so there is something to refresh.
	writeClaudeFixture(t, filepath.Join(tmp, "claude", "app", "two.jsonl"), "s2", []string{
		`{"type":"user","sessionId":"s2","timestamp":"2026-09-01T11:00:00Z","message":{"role":"user","content":"the cache key ignored the tenant id"}}`,
	})
	holdTheIndexLock(t)

	var said bytes.Buffer
	if err := ensureForCLISearch(index.DefaultDir(), query.Options{Query: "retry"}, false, &said); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if !strings.Contains(said.String(), "as it was") {
		t.Fatalf("a store that can answer said nothing about serving the older view:\n%s", said.String())
	}
}

// holdTheIndexLock takes the lock the way another deja would, so the search
// under test takes the path it takes when a build is running.
func holdTheIndexLock(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the lock is taken differently on windows")
	}
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
