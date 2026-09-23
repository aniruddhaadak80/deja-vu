package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The opencode plugin must fold its recall into the first system block. An
// OpenAI-compatible endpoint that requires the system message to come first
// rejects a request carrying a second one, and installing deja then made every
// turn fail: "Not Found: System message must be at the beginning." Reproduced
// against a local model, and fixed by merging rather than appending.
func TestOpencodePluginDoesNotAppendASecondSystemBlock(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	// The shape is picked from the installed opencode, and the test is about
	// the 2.x one.
	t.Setenv("DEJA_OPENCODE_MAJOR", "2")

	if _, err := installOpencodePlugin("/usr/local/bin/deja", false); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".config", "opencode", "plugins", "deja.js")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	src := string(b)
	compact := strings.Join(strings.Fields(src), "")
	if !strings.Contains(compact, `if(event.system.length)event.system[0].text=digest+"\n\n"+event.system[0].text`) {
		t.Fatal("plugin no longer folds the recall into the first system block")
	}
	// The push only runs when there is no system block at all, so a second
	// block never lands behind an existing one.
	if !strings.Contains(compact, `elseevent.system.push({type:"text",text:digest})`) {
		t.Fatal("the push is not guarded by emptiness")
	}
}
