//go:build windows

package main

import "testing"

// Windows takes the index lock through the share mode of the open itself, and a
// second opener in the same process is not what that refuses. The tests that
// need a lock somebody else holds are skipped rather than faked.
func holdTheIndexLock(t *testing.T) {
	t.Helper()
	t.Skip("the index lock is taken differently on windows")
}
