// Package corpus is intentionally flawed Go code used to generate a REAL
// staticcheck JSON corpus (see detectors_test.go: corpus files are captured by
// running `staticcheck -f json ./...` in this directory; the captured JSONL is
// committed as staticcheck-corpus.jsonl next to this file). Never fix these
// findings — they are the fixture.
package corpus

import (
	"fmt"
	"strings"
)

// Compare triggers S1002 (boolean comparison with literal).
func Compare(b bool) bool {
	return b == true
}

// NeverRead triggers SA4006 (assigned value never read).
func NeverRead(x int) int {
	x = x + 1

	return 0
}

// SprintfWrap triggers S1025 (argument is already a string).
func SprintfWrap(s string) string {
	return fmt.Sprintf("%s", s)
}

// ReplaceAll triggers S1003 (strings.Replace count argument should be -1).
func ReplaceAll(s string) string {
	return strings.Replace(s, "a", "b", -1)
}
