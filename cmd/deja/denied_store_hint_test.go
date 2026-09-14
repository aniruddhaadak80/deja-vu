package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// A store deja can see and cannot open. #1020 wrote the sentence for it, inside
// the branch for a machine with no history at all — and a store whose files are
// visible counts as history, so the sentence was unreachable in exactly the case
// it was written for. What the reader got instead was "run `deja index`", which
// is what they had just done and which cannot help (#3585).
func TestASearchOverADeniedStoreSaysSoRatherThanAdvisingAnIndex(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod 000 does not deny a read on windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root reads a file whatever its mode")
	}
	tmp := hermeticEnv(t)
	transcript := filepath.Join(tmp, "claude", "-work-app", "u1.jsonl")
	writeClaudeFixture(t, transcript, "u1", []string{
		`{"type":"user","sessionId":"u1","timestamp":"2026-09-01T10:00:00Z","cwd":"/work/app","message":{"role":"user","content":"the exporter drops every third retry"}}`,
	})
	if err := os.Chmod(transcript, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(transcript, 0o644) })

	hint := emptyIndexHint("no matches for \"exporter\"")
	if strings.Contains(hint, "run `deja index`") {
		t.Fatalf("a denied store was told to run the index it just ran: %q", hint)
	}
	if !strings.Contains(hint, "permission denied") {
		t.Fatalf("hint = %q, want it to name the wall", hint)
	}
	if !strings.Contains(hint, "deja doctor") {
		t.Fatalf("hint = %q, want it to point at the command that names the path", hint)
	}
}

// The two sentences either side of it are unchanged: a machine with nothing,
// and a machine whose store simply has not been indexed.
func TestTheOtherEmptyHintsAreUnchanged(t *testing.T) {
	tmp := hermeticEnv(t)
	if hint := emptyIndexHint("no matches"); !strings.Contains(hint, "no agent history was found") {
		t.Errorf("an empty machine = %q", hint)
	}
	writeClaudeFixture(t, filepath.Join(tmp, "claude", "-work-app", "u1.jsonl"), "u1", []string{
		`{"type":"user","sessionId":"u1","timestamp":"2026-09-01T10:00:00Z","cwd":"/work/app","message":{"role":"user","content":"the exporter drops every third retry"}}`,
	})
	if hint := emptyIndexHint("no matches"); !strings.Contains(hint, "run `deja index`") {
		t.Errorf("a readable but unindexed store = %q", hint)
	}
}
