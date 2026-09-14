package main

import (
	"strings"
	"testing"
)

// The last slice of a session has no next one. The note pointed at the offset
// after it anyway, and following that advice answered "--offset 16 is past the
// end — the session has 16 messages": deja arguing with itself (#3578).
func TestTheLastSliceDoesNotOfferANextOne(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		offset, returned, total int
		wantNext                bool
		says                    string
	}{
		{name: "a middle slice", offset: 6, returned: 3, total: 16, wantNext: true, says: "7-9 of 16"},
		{name: "the last slice", offset: 13, returned: 3, total: 16, says: "14-16 of 16"},
		{name: "a slice that ends exactly", offset: 14, returned: 2, total: 16, says: "15-16 of 16"},
		{name: "the whole session at an offset", offset: 1, returned: 15, total: 16, says: "2-16 of 16"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			note := showWindowNote(tc.offset, tc.returned, tc.total)
			if !strings.Contains(note, tc.says) {
				t.Fatalf("note = %q, want it to say %q", note, tc.says)
			}
			hasNext := strings.Contains(note, "reads the next slice")
			if hasNext != tc.wantNext {
				t.Fatalf("note = %q, offering a next slice = %v, want %v", note, hasNext, tc.wantNext)
			}
			if !tc.wantNext && !strings.Contains(note, "end of the session") {
				t.Fatalf("note = %q, want it to say where the reader is", note)
			}
		})
	}
}

// The cases either side of it keep their own sentences.
func TestTheOtherWindowNotesAreUnchanged(t *testing.T) {
	if note := showWindowNote(0, 16, 16); note != "" {
		t.Errorf("a reader who asked for everything got arithmetic: %q", note)
	}
	if note := showWindowNote(0, 0, 0); !strings.Contains(note, "no messages to show") {
		t.Errorf("an empty session = %q", note)
	}
	if note := showWindowNote(16, 0, 16); !strings.Contains(note, "past the end") {
		t.Errorf("past the end = %q", note)
	}
}
