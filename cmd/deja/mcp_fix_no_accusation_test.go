package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/vshulcz/deja-vu/internal/index"
)

// `fix` used the pair-mining heuristic on the caller: text it did not recognise
// was answered "that does not read like an error line - pass the failing output
// itself", which is an accusation, and a wrong one. Measured against twenty
// error lines a tool actually prints, ten were refused — `connection refused`,
// `permission denied`, `segmentation fault`, `OOMKilled`, `Exit code 137` among
// them — while the CLI took all twenty (#3580).
func TestFixDoesNotTellAnAgentItsErrorIsNotAnError(t *testing.T) {
	hermeticEnv(t)
	for _, text := range []string{
		"connection refused",
		"permission denied",
		"segmentation fault",
		"OOMKilled",
		"Exit code 137",
		"undefined: foo",
		"context deadline exceeded",
		// And the case the guidance is for: prose about a failure rather than
		// the failure.
		"it did not work",
	} {
		out, _, err := mcpFix(index.DefaultDir(), "deja", mustFixArgs(t, text))
		if err != nil {
			t.Fatalf("%q: %v", text, err)
		}
		if strings.Contains(out, "does not read like an error") {
			t.Errorf("%q was told it is not an error: %q", text, out)
		}
		if !strings.Contains(out, "No session on this machine ran a command after that error") {
			t.Errorf("%q got no honest answer: %q", text, out)
		}
	}
}

// The advice itself is worth keeping, so it rides along where the heuristic
// says the text may be a summary.
func TestFixStillSuggestsPastingTheOutputWhenTheTextLooksLikeProse(t *testing.T) {
	hermeticEnv(t)
	out, _, err := mcpFix(index.DefaultDir(), "deja", mustFixArgs(t, "it did not work"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "pass the output itself") {
		t.Fatalf("the guidance is gone: %q", out)
	}
	// And it does not ride along where the text is plainly an error already.
	out, _, err = mcpFix(index.DefaultDir(), "deja", mustFixArgs(t, "fatal: not a git repository (or any of the parent directories): .git"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "pass the output itself") {
		t.Fatalf("an error line was offered advice about summaries: %q", out)
	}
}

func mustFixArgs(t *testing.T, text string) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(map[string]any{"error": text})
	if err != nil {
		t.Fatal(err)
	}
	return b
}
