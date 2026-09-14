package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vshulcz/deja-vu/internal/index"
)

// The text screen has carried a `format` row since #877 and the JSON carried
// nothing, so a script watching index health read "ok" on a store whose recall
// was off — the same miss #2292 closed for damage (#3600).
func TestDoctorJSONSaysWhenTheStoreIsNotWhatThisBuildWrites(t *testing.T) {
	tmp := hermeticEnv(t)
	writeClaudeFixture(t, filepath.Join(os.Getenv("DEJA_CLAUDE_ROOT"), "work", "one.jsonl"), "fmt", []string{
		`{"type":"user","sessionId":"fmt","timestamp":"2026-08-20T10:00:00Z","cwd":"/work/app",` +
			`"message":{"role":"user","content":"the migration ran twice"}}`,
	})
	dir := filepath.Join(tmp, "index.db")
	if err := index.Ensure(dir, "", false, nil); err != nil {
		t.Fatal(err)
	}

	// A store this build wrote says nothing about its format.
	got := inspectDoctorIndex(dir, nil)
	if got.Format != "" {
		t.Errorf("a current store reported a format: %q", got.Format)
	}
	if got.State != "ok" {
		t.Errorf("a current store is not ok: %q", got.State)
	}

	saved := indexReadState
	t.Cleanup(func() { indexReadState = saved })
	for _, c := range []struct {
		state        index.ReadState
		wantFormat   string
		wantState    string
		healthSignal string
	}{
		{index.ReadStateWithheld, "withheld", "rereading", "recall is off"},
		{index.ReadStateUnreadable, "unreadable", "rereading", "recall is off"},
		{index.ReadStateOlderRules, "older-rules", "ok", "still answering"},
		{index.ReadStateNewer, "newer", "ok", "still answering"},
	} {
		indexReadState = func(string) index.ReadState { return c.state }
		got := inspectDoctorIndex(dir, nil)
		if got.Format != c.wantFormat {
			t.Errorf("%s: format = %q, want %q", c.healthSignal, got.Format, c.wantFormat)
		}
		if got.State != c.wantState {
			t.Errorf("%s: state = %q, want %q", c.healthSignal, got.State, c.wantState)
		}
	}

	// Every value the field can take has to be named in the contract, or a
	// consumer branching on it is reading prose that does not cover it.
	doc, err := os.ReadFile("../../docs/json-output.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"unreadable", "withheld", "older-rules", "newer", "rereading"} {
		if !strings.Contains(string(doc), "`"+value+"`") {
			t.Errorf("docs/json-output.md does not name the %q state", value)
		}
	}
}
