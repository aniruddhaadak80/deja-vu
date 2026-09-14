package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A store root deja was *told* about is not a store that is there. Antigravity
// is the one harness whose roots come back unchecked — the env override is
// returned as given — so a machine with the variable pointing at a directory
// that does not exist reported `found` with nothing in it, where every other
// row says `missing`. doctor's whole job is saying which of those it is.
func TestDoctorSaysMissingForAnAntigravityRootThatIsNotThere(t *testing.T) {
	tmp := hermeticEnv(t)
	t.Setenv("DEJA_ANTIGRAVITY_ROOT", filepath.Join(tmp, "no-such-antigravity"))

	var buf bytes.Buffer
	if err := runDoctor(&buf, []string{"--offline"}, stubLookup("1.0.0", false), t.TempDir()); err != nil {
		t.Fatal(err)
	}
	row := harnessRow(t, buf.String(), "antigravity")
	if !strings.Contains(row, "missing") {
		t.Fatalf("row = %q — a root that is not on disk is missing, not found", row)
	}

	// And it flips the moment the directory is there, so this is about the
	// store rather than about the variable.
	root := filepath.Join(tmp, "antigravity-here")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DEJA_ANTIGRAVITY_ROOT", root)
	buf.Reset()
	if err := runDoctor(&buf, []string{"--offline"}, stubLookup("1.0.0", false), t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if row := harnessRow(t, buf.String(), "antigravity"); !strings.Contains(row, "found") {
		t.Fatalf("row = %q — the directory exists, so the store is there", row)
	}
}
