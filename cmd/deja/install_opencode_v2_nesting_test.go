package main

import (
	"strings"
	"testing"
)

// OpenCode 2.x keeps its MCP servers under `mcp.servers`, one level deeper
// than the 1.x shape this text writer edits. The writer found deja's entry
// by matching the key at any depth and re-inserted it at the top of the
// `mcp` block, so a 2.x config came back with deja moved out of
// `mcp.servers` and the reader's `environment` block stranded beside the
// now-empty `servers` — and the install reported a normal replacement
// (#3929). deja must refuse to edit this nesting rather than relocate the
// entry to a place the harness no longer reads.
func TestOpencodeJSONCRefusesV2NestedServers(t *testing.T) {
	seed := "{\n" +
		"  // my MCP servers\n" +
		"  \"mcp\": {\n" +
		"    \"servers\": {\n" +
		"      \"deja\": {\n" +
		"        \"type\": \"local\",\n" +
		"        \"command\": [\"/usr/local/bin/deja\", \"mcp\"],\n" +
		"        \"environment\": { \"DEJA_HOME\": \"/mnt/nas/deja\" }\n" +
		"      }\n" +
		"    }\n" +
		"  }\n" +
		"}\n"

	out, _, err := updateOpencodeJSONC([]byte(seed), "/bin/deja", false)
	if err == nil {
		t.Fatalf("install edited a V2-nested config instead of refusing;\n%s", out)
	}
	if !strings.Contains(err.Error(), "mcp.servers") {
		t.Errorf("the refusal does not name the nesting it cannot edit: %v", err)
	}
}

// The 1.x shape — deja directly under `mcp` — is what this writer speaks,
// and the fix must not refuse it.
func TestOpencodeJSONCStillEditsV1TopLevelServers(t *testing.T) {
	seed := "{\n" +
		"  // my MCP servers\n" +
		"  \"mcp\": {\n" +
		"    \"deja\": {\"type\":\"local\",\"command\":[\"/old/deja\",\"mcp\"]}\n" +
		"  }\n" +
		"}\n"

	out, _, err := updateOpencodeJSONC([]byte(seed), "/bin/deja", false)
	if err != nil {
		t.Fatalf("the 1.x shape was refused: %v", err)
	}
	if !strings.Contains(string(out), "/bin/deja") {
		t.Errorf("the binary was not updated:\n%s", out)
	}
}

// The same 2.x nesting written inline — the whole `servers` object on one
// line — is the same hazard: the drop matches the key wherever it sits, so
// it would take the whole `servers` line out, deleting the reader's other
// servers with it, and write deja at the top of `mcp`. The guard must see
// the nested key mid-line, not only at a line start.
func TestOpencodeJSONCRefusesInlineV2NestedServers(t *testing.T) {
	seed := "{\n" +
		"  \"mcp\": {\n" +
		"    \"servers\": { \"deja\": {\"type\":\"local\",\"command\":[\"/old/deja\",\"mcp\"]}, \"theirs\": {\"command\":[\"other\"]} }\n" +
		"  }\n" +
		"}\n"

	out, _, err := updateOpencodeJSONC([]byte(seed), "/bin/deja", false)
	if err == nil {
		t.Fatalf("install edited an inline V2-nested config instead of refusing;\n%s", out)
	}
	if !strings.Contains(err.Error(), "mcp.servers") {
		t.Errorf("the refusal does not name the nesting it cannot edit: %v", err)
	}
}
