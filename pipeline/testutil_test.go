package pipeline

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
)

type mockDetector struct {
	name     string
	findings []finding.Finding
	err      error
	delay    time.Duration
}

func (m *mockDetector) Name() string {
	return m.name
}

func (m *mockDetector) Detect(ctx context.Context) ([]finding.Finding, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return m.findings, m.err
}

func newMockDetector(name, toolName string, findings ...finding.Finding) *mockDetector {
	if len(findings) == 0 {
		findings = []finding.Finding{
			{
				ID:       "1",
				Rule:     "r1",
				ToolName: toolName,
				Message:  "m",
				Severity: finding.SeverityError,
			},
		}
	}

	return &mockDetector{name: name, findings: findings}
}

func testFinding(id, rule, tool, msg string, sev finding.Severity, file string) finding.Finding {
	return finding.Finding{
		ID:       id,
		Rule:     rule,
		ToolName: tool,
		Message:  msg,
		Severity: sev,
		Position: finding.Position{File: file},
	}
}

func findingWithRange(id, file string, line, startLine, endLine int) finding.Finding {
	return finding.Finding{
		ID:       id,
		Position: finding.Position{File: file, Line: line},
		Range: &finding.Range{
			Start: finding.Position{File: file, Line: startLine},
			End:   finding.Position{File: file, Line: endLine},
		},
	}
}

func findingAt(id, file string, line int) finding.Finding {
	return finding.Finding{ID: id, Position: finding.Position{File: file, Line: line}}
}

func findings(fixSpecs ...any) []finding.Finding {
	var fixes []finding.Finding

	for i := 0; i+2 < len(fixSpecs); i += 3 {
		id, idOk := fixSpecs[i].(string)
		file, fileOk := fixSpecs[i+1].(string)

		line, lineOk := fixSpecs[i+2].(int)
		if idOk && fileOk && lineOk {
			fixes = append(fixes, findingAt(id, file, line))
		}
	}

	return fixes
}

func overlappingFindings() []finding.Finding {
	return []finding.Finding{
		findingWithRange("1", "a.go", 10, 10, 20),
		findingWithRange("2", "a.go", 15, 15, 25),
	}
}

func assertFindingsFound(t *testing.T, result *PipelineResult, want int, msg string) {
	t.Helper()

	if len(result.Iterations) == 0 {
		t.Fatal("expected at least one iteration")
	}

	found := result.Iterations[0].FindingsFound
	if found != want {
		t.Errorf("%s = %d, want %d", msg, found, want)
	}
}

func assertIterationsLen(t *testing.T, result *PipelineResult, want int) {
	t.Helper()

	if n := len(result.Iterations); n != want {
		t.Fatalf("iterations = %d, want %d", n, want)
	}
}

func requireNoError(t *testing.T, err error, msg string) {
	t.Helper()

	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

func requireLenEq(t *testing.T, got, want int, msg string) {
	t.Helper()

	if got != want {
		t.Fatalf("%s: got %d, want %d", msg, got, want)
	}
}

func assertIntEq(t *testing.T, got, want int, field string) {
	t.Helper()

	if got != want {
		t.Errorf("%s = %d, want %d", field, got, want)
	}
}

func writeFile(path string, data []byte, perm uint32) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(perm))
	if err != nil {
		return err
	}

	_, err = f.Write(data)

	err1 := f.Close()
	if err1 != nil && err == nil {
		err = err1
	}

	return err
}

func writeTestFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := writeFile(path, data, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func readFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() { _ = f.Close() }()

	var result []byte

	buf := make([]byte, 1024)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			result = append(result, buf[:n]...)
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, err
		}
	}

	return result, nil
}

func testBackupRestore(t *testing.T, applier *FixApplier, original, modified string) {
	t.Helper()

	testFile := filepath.Join(t.TempDir(), "roundtrip.go")
	writeTestFile(t, testFile, []byte(original))

	if err := applier.backup(testFile); err != nil {
		t.Fatalf("backup: %v", err)
	}

	if err := writeFile(testFile, []byte(modified), 0o644); err != nil {
		t.Fatalf("write modified: %v", err)
	}

	if err := applier.restore(testFile); err != nil {
		t.Fatalf("restore: %v", err)
	}

	data, err := readFile(testFile)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if string(data) != original {
		t.Errorf("restored content mismatch:\ngot:  %q\nwant: %q", string(data), original)
	}
}

func makeFixFinding(id, before, after, file string, line int) finding.Finding {
	return finding.Finding{
		ID:          id,
		Rule:        "r1",
		ToolName:    "tool",
		Message:     "replace " + before + " with " + after,
		BeforeCode:  before,
		AfterCode:   after,
		Position:    finding.Position{File: file, Line: line},
		FixStrategy: finding.FixStrategyDirect,
	}
}

func makeErrorDetector(errMsg string) Detector {
	return DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
		return nil, errors.New(errMsg)
	})
}

func assertFindingErrorIO(t *testing.T, fe *finding.FindingError, file string) {
	t.Helper()

	if fe.Category != finding.ErrCategoryIO {
		t.Errorf("category = %q, want %q", fe.Category, finding.ErrCategoryIO)
	}

	if fe.Position.File != file {
		t.Errorf("file = %q, want %q", fe.Position.File, file)
	}
}

func makeFixFindingWithRange(id, before, after, file string, line, col int) finding.Finding {
	return finding.Finding{
		ID:          id,
		BeforeCode:  before,
		AfterCode:   after,
		Position:    finding.Position{File: file, Line: line, Column: col},
		Range:       finding.NewRangePtr(file, line, col, line, col+len(before)),
		FixStrategy: finding.FixStrategyDirect,
	}
}

func makeConflictingFixes() []finding.Finding {
	return []finding.Finding{
		makeFixFindingWithRange("fix1", "old()", "new()", "fixme.go", 4, 2),
		makeFixFindingWithRange("fix2", "old()", "other()", "fixme.go", 4, 2),
	}
}

var suppressedIDs = []string{"s2", "s4"}

func assertNoSuppressedFindings(t *testing.T, findings []finding.Finding, mode string) {
	t.Helper()
	for _, f := range findings {
		for _, sid := range suppressedIDs {
			if f.ID == sid {
				t.Errorf("suppressed finding %s present in %s results", f.ID, mode)
			}
		}
	}
}

func assertNoSuppressedNotified(t *testing.T, notified []string, mode string) {
	t.Helper()
	for _, id := range notified {
		for _, sid := range suppressedIDs {
			if id == sid {
				t.Errorf("OnFinding called for suppressed finding %s in %s mode", id, mode)
			}
		}
	}
}
