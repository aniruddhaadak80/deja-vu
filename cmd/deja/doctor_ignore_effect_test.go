package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/vshulcz/deja-vu/internal/index"
	"github.com/vshulcz/deja-vu/internal/policy"
)

// An ignore rule is matched against the project name and the transcript's own
// path, so the natural thing to write — the directory's absolute path — matches
// neither, and doctor reported it as in force while it hid nothing. The
// activation rows above it have said what they withhold since #978; this is the
// row that did not (#3584).
func TestDoctorSaysWhatAnIgnoreRuleActuallyHides(t *testing.T) {
	tmp := hermeticEnv(t)
	writeClaudeFixture(t, tmp+"/claude/-Users-me-code-client-work/c1.jsonl", "c1", []string{
		`{"type":"user","sessionId":"c1","timestamp":"2026-09-01T10:00:00Z","cwd":"/Users/me/code/client-work","message":{"role":"user","content":"the exporter drops every third retry"}}`,
	})
	writeClaudeFixture(t, tmp+"/claude/-Users-me-code-open/o1.jsonl", "o1", []string{
		`{"type":"user","sessionId":"o1","timestamp":"2026-09-01T10:00:00Z","cwd":"/Users/me/code/open","message":{"role":"user","content":"the cache key ignored the tenant id"}}`,
	})
	if _, err := captureRunStderr(t, "index"); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name, pattern, says string
	}{
		{
			name:    "a rule that matches nothing",
			pattern: "/Users/me/code/client-work",
			says:    "matches no indexed session",
		},
		{
			name:    "a rule that hides something",
			pattern: "client-work",
			says:    "hides 1 of 2 indexed sessions",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			printIgnored(&out, policy.Policy{Ignore: []string{tc.pattern}}, index.DefaultDir())
			if !strings.Contains(out.String(), tc.pattern) {
				t.Fatalf("the rule is not named:\n%s", out.String())
			}
			if !strings.Contains(out.String(), tc.says) {
				t.Fatalf("output = %q, want it to say %q", out.String(), tc.says)
			}
		})
	}
}

// deja's own default is not something to explain to a machine that has never
// met an agent runtime: only a rule somebody wrote gets the effect line.
func TestTheDefaultIgnoreRuleIsNotAnnotated(t *testing.T) {
	hermeticEnv(t)
	var out bytes.Buffer
	printIgnored(&out, policy.Policy{}, index.DefaultDir())
	if out.Len() == 0 {
		t.Fatal("the default rule is not printed at all")
	}
	if strings.Contains(out.String(), "matches no indexed session") || strings.Contains(out.String(), "hides") {
		t.Fatalf("the default rule was annotated:\n%s", out.String())
	}
}
