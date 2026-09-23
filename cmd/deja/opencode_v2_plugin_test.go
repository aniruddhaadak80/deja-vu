package main

import (
	"strings"
	"testing"
)

// OpenCode V2 reads only the module's default export, and that export must
// carry an id and a setup (or effect) function. The V1 named-export shape —
// `export const DejaRecall = async ({ $, client, directory }) => ...` — left
// the loader with no default at all, and every load failed with
// "Plugin must export a default definition with an id and an effect or setup
// function. (cause: SchemaError(Missing key at ["default"]))" (ref
// err_bb76d60a). The plugin was installed and silently did nothing.
func TestOpencodePluginExportsV2Definition(t *testing.T) {
	js := opencodePluginJS("/bin/deja")
	compact := strings.Join(strings.Fields(js), "")

	for _, want := range []string{
		"exportdefault{",
		`id:"deja-recall"`,
		"asyncsetup(ctx)",
	} {
		if !strings.Contains(compact, want) {
			t.Errorf("generated plugin missing %q — the V2 loader rejects the module without it:\n%s", want, js)
		}
	}

	// The plugin must not import @opencode/plugin: a local plugin resolves
	// bare specifiers from its own directory, where the package is not
	// installed ("Cannot find package '@opencode/plugin' imported from
	// .../plugins/deja.js"). Plugin.define is the identity function, so the
	// plain { id, setup } object is the same contract with no dependency.
	if strings.Contains(compact, `from"@opencode/plugin"`) {
		t.Error("the plugin imports @opencode/plugin, which a local plugin cannot resolve")
	}

	// The V1 named export is exactly what the V2 loader cannot see. If it
	// rides along, a future reader may think the plugin still speaks V1.
	if strings.Contains(compact, "exportconstDejaRecall") {
		t.Error("the V1 named export is still there; the V2 loader only reads the default export")
	}

	// The digest still arrives in the recall envelope, so a rename on either
	// side is caught here rather than in a silently empty context.
	if !strings.Contains(compact, "hookSpecificOutput") {
		t.Error("the plugin no longer reads the recall envelope")
	}
}
