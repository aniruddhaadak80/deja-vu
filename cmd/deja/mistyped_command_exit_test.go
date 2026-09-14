package main

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A search that finds nothing is the ordinary answer to a question the history
// cannot settle, and it exits 0. A command name that does not exist is not:
// `deja unforget x` did nothing, said so, and exited 0 — so a script that ran
// it and checked the code was told the work had been done.
func TestAMistypedCommandExitsNonZeroAndAMissDoesNot(t *testing.T) {
	seedOneSession(t)

	for _, tc := range []struct {
		name    string
		args    []string
		wantErr bool
		says    string
	}{
		{
			name:    "a command nobody spelled right",
			args:    []string{"unforget", "nope"},
			wantErr: true,
			says:    "`deja forget --unforget <id>`",
		},
		{
			name:    "a command one letter off",
			args:    []string{"doctro"},
			wantErr: true,
			says:    "did you mean `deja doctor`",
		},
		{
			name: "a search that found nothing",
			args: []string{"zzzqqq-nothing-here"},
			says: "no matches",
		},
		{
			name: "a search that found something",
			args: []string{"exporter"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			said, err := captureRunStderr(t, tc.args...)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, want error = %v (stderr %q)", err, tc.wantErr, said)
			}
			if tc.wantErr && !errors.Is(err, errAlreadySaid) {
				t.Fatalf("err = %v, want the one that exits without saying anything more", err)
			}
			if tc.says != "" && !strings.Contains(said, tc.says) {
				t.Fatalf("stderr = %q, want it to say %q", said, tc.says)
			}
		})
	}
}

// The sentence is printed once, by the command rather than by main: the exit
// code is all errAlreadySaid adds.
func TestTheMistypedCommandSentenceIsNotPrintedTwice(t *testing.T) {
	seedOneSession(t)
	said, err := captureRunStderr(t, "doctro")
	if err == nil {
		t.Fatal("a mistyped command exited 0")
	}
	if n := strings.Count(said, "is not a command"); n != 1 {
		t.Fatalf("the sentence appears %d times:\n%s", n, said)
	}
	if strings.Contains(said, "already said") {
		t.Fatalf("the sentinel reached the screen:\n%s", said)
	}
}

// A typo whose word is in the history is a different case: the search answered,
// so the run succeeded and the hint is a footnote on stderr (#2197).
func TestAMistypedCommandThatMatchedSomethingStillExitsZero(t *testing.T) {
	tmp := hermeticEnv(t)
	claude := filepath.Join(tmp, "claude")
	t.Setenv("DEJA_CLAUDE_ROOT", claude)
	at := time.Now().Add(-72 * time.Hour).UTC().Format(time.RFC3339)
	writeClaudeFixture(t, filepath.Join(claude, "alpha", "one.jsonl"), "c1", []string{
		`{"type":"user","sessionId":"c1","timestamp":"` + at +
			`","message":{"role":"user","content":"we ran deja stats and doctro on the box"}}`,
	})
	if _, err := captureRunStderr(t, "index"); err != nil {
		t.Fatal(err)
	}
	said, err := captureRunStderr(t, "doctro")
	if err != nil {
		t.Fatalf("err = %v — the search answered, so the run worked", err)
	}
	if !strings.Contains(said, "did you mean `deja doctor`?") {
		t.Fatalf("the hint is gone:\n%s", said)
	}
}

func seedOneSession(t *testing.T) {
	t.Helper()
	tmp := hermeticEnv(t)
	claude := filepath.Join(tmp, "claude")
	t.Setenv("DEJA_CLAUDE_ROOT", claude)
	at := time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339)
	writeClaudeFixture(t, filepath.Join(claude, "app", "one.jsonl"), "s1", []string{
		`{"type":"user","sessionId":"s1","timestamp":"` + at +
			`","message":{"role":"user","content":"the exporter drops every third retry"}}`,
	})
	if _, err := captureRunStderr(t, "index"); err != nil {
		t.Fatal(err)
	}
}
