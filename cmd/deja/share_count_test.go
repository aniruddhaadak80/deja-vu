package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// The line under a share tells the user how much was scrubbed, and it read
// "0 secrets masked" on a document visibly full of markers: secrets are
// redacted at index time, so the pass share runs over already-clean text
// replaces nothing.
func TestShareCountsSecretsRedactedEarlier(t *testing.T) {
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	// Drained while the call runs: a pipe holds one buffer, and a capture that
	// reads after the call deadlocks as soon as the output outgrows it.
	drained := drainPipe(r)
	var out bytes.Buffer
	printSanitized(&out, "key [redacted:openai-key] and password [redacted:credential]\n")
	_ = w.Close()
	os.Stderr = old
	if msg := <-drained; !strings.Contains(msg, "2 secrets masked") {
		t.Fatalf("count is wrong: %q", msg)
	}
}

func TestShareCountsZeroWhenNothingWasRedacted(t *testing.T) {
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	drained := drainPipe(r)
	var out bytes.Buffer
	printSanitized(&out, "nothing sensitive here at all\n")
	_ = w.Close()
	os.Stderr = old
	if msg := <-drained; !strings.Contains(msg, "0 secrets masked") {
		t.Fatalf("clean document reported as masked: %q", msg)
	}
}

// drainPipe reads the pipe's reader end in a goroutine and hands the text back
// on a channel, which is the only safe order: the writer is the code under
// test, and it blocks once the buffer fills — 4 KB on windows, where the leg
// hung on main rather than on the pull request that grew the output (#3493).
func drainPipe(r *os.File) <-chan string {
	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	return done
}
