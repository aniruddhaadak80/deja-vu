package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vshulcz/deja-vu/internal/index"
)

// The environment block reads the friction sidecar rather than the records, so
// it went around the gate that stops a store below the redaction floor from
// being quoted — and put a wall from the old store, secret and all, straight
// into the model's context on the first session after an upgrade. Every other
// surface is safe because it runs index.Ensure first; the session-start hook
// cannot, which is the whole reason the floor exists (#3598).
func TestTheEnvironmentBlockIsQuietWhileTheStoreIsBelowTheFloor(t *testing.T) {
	tmp := hermeticEnv(t)
	root := os.Getenv("DEJA_CLAUDE_ROOT")
	wall := "dial tcp 127.0.0.1:5432: connect: connection refused"
	for _, id := range []string{"w1", "w2", "w3", "w4"} {
		writeClaudeFixture(t, filepath.Join(root, "work", id+".jsonl"), id, []string{
			`{"type":"user","sessionId":"` + id + `","timestamp":"2026-09-01T09:00:00Z","cwd":"/work/app",` +
				`"message":{"role":"user","content":"the api will not start"}}`,
			`{"type":"assistant","sessionId":"` + id + `","timestamp":"2026-09-01T09:00:01Z","cwd":"/work/app",` +
				`"message":{"role":"assistant","content":[{"type":"tool_use","name":"Bash","input":{"command":"go test ./..."}}]}}`,
			`{"type":"user","sessionId":"` + id + `","timestamp":"2026-09-01T09:00:02Z","cwd":"/work/app",` +
				`"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"t1","is_error":true,"content":"` + wall + `"}]}}`,
		})
	}
	dir := filepath.Join(tmp, "index.db")
	if err := index.Ensure(dir, "", false, nil); err != nil {
		t.Fatal(err)
	}

	// The store this build wrote answers, or the test below proves nothing.
	if block := environmentBlock(dir, "auto"); !strings.Contains(block, wall) {
		t.Fatalf("the block says nothing on a current store:\n%s", block)
	}

	// And a store whose text predates a redaction fix says nothing at all,
	// rather than the wall it recorded under the older rules.
	saved := environmentReadState
	environmentReadState = func(string) index.ReadState { return index.ReadStateWithheld }
	t.Cleanup(func() { environmentReadState = saved })
	if block := environmentBlock(dir, "auto"); block != "" {
		t.Errorf("a store below the redaction floor was quoted:\n%s", block)
	}

	environmentReadState = func(string) index.ReadState { return index.ReadStateUnreadable }
	if block := environmentBlock(dir, "auto"); block != "" {
		t.Errorf("a store this build cannot read was quoted:\n%s", block)
	}

	// A store one derivation behind still answers: those bumps outnumber the
	// redaction ones three to one, and going quiet on them would take the block
	// off most upgrades for nothing.
	environmentReadState = func(string) index.ReadState { return index.ReadStateOlderRules }
	if block := environmentBlock(dir, "auto"); !strings.Contains(block, wall) {
		t.Errorf("a readable store lost the block:\n%s", block)
	}
}
