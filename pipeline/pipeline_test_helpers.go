package pipeline

import "github.com/larsartmann/go-finding"

// conflictTestCase represents a conflict detection test case.
type conflictTestCase struct {
	name              string
	fixes             []finding.Finding
	expectedGroups    int
	expectedConflicts int
}

// testingT is a subset of testing.T for helper functions.
type testingT interface {
	Helper()
	Errorf(format string, args ...interface{})
}

// assertFindingsCount checks that findings have expected length.
func assertFindingsCount(t testingT, findings []finding.Finding, expected int) {
	if len(findings) != expected {
		t.Errorf("expected %d findings, got %d", expected, len(findings))
	}
}
