package pipeline

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/larsartmann/go-finding"
	"github.com/onsi/gomega"
)

// NewParallelGomega marks the test as parallel and returns a gomega.GomegaWithT,
// consolidating the t.Parallel() + gomega.NewWithT(t) boilerplate.
func NewParallelGomega(t *testing.T) *gomega.GomegaWithT {
	t.Helper()
	t.Parallel()

	return gomega.NewWithT(t)
}

const helloProgram = "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"

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
				ToolName: finding.ToolName(toolName),
				Message:  "m",
				Severity: finding.SeverityError,
			},
		}
	}

	return &mockDetector{name: name, findings: findings}
}

func mockDet(name string, findingID finding.ID) *mockDetector {
	return &mockDetector{name: name, findings: []finding.Finding{{ID: findingID}}}
}

func mockDetWithFindings(name string, findings ...finding.Finding) *mockDetector {
	return &mockDetector{name: name, findings: findings}
}

func slowMockDetector(name string, delay time.Duration, findings ...finding.Finding) *mockDetector {
	return &mockDetector{name: name, delay: delay, findings: findings}
}

func slowTestDetector(name string, delay time.Duration) *mockDetector {
	return slowMockDetector(
		name, delay,
		finding.Finding{ID: "t:r:f:1", Rule: "r", ToolName: "t", Message: "m"},
	)
}

func testFinding(id, rule, tool, msg string, sev finding.Severity, file string) finding.Finding {
	return finding.Finding{
		ID:       finding.ID(id),
		Rule:     finding.RuleName(rule),
		ToolName: finding.ToolName(tool),
		Message:  msg,
		Severity: sev,
		Position: finding.Position{File: finding.FilePath(file)},
	}
}

// registerFindingDetector registers a single-finding detector factory with the
// given registry under `name`, producing findings tagged with the provided
// `tool`. Both fields that the mocks and production detectors care about.
func registerFindingDetector(registry *finding.DetectorRegistry, name, tool string, f finding.Finding) {
	registry.MustRegister(name, func() finding.Detector {
		return newMockDetector(name, tool, f)
	})
}

func findingWithRange(id, file string, line, startLine, endLine int) finding.Finding {
	return finding.Finding{
		ID:       finding.ID(id),
		Position: finding.Position{File: finding.FilePath(file), Line: line},
		Range: &finding.Range{
			Start: finding.Position{File: finding.FilePath(file), Line: startLine},
			End:   finding.Position{File: finding.FilePath(file), Line: endLine},
		},
	}
}

func findingAt(id, file string, line int) finding.Finding {
	return finding.Finding{ID: finding.ID(id), Position: finding.Position{File: finding.FilePath(file), Line: line}}
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

	err := writeFile(path, data, 0o644)
	if err != nil {
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

	err := applier.backup.Backup(testFile)
	if err != nil {
		t.Fatalf("backup: %v", err)
	}

	err = writeFile(testFile, []byte(modified), 0o644)
	if err != nil {
		t.Fatalf("write modified: %v", err)
	}

	err = applier.backup.Restore(testFile)
	if err != nil {
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

func newTestApplier(t *testing.T) *FixApplier {
	t.Helper()

	a, err := NewFixApplier(t.TempDir())
	if err != nil {
		t.Fatalf("NewFixApplier: %v", err)
	}

	return a
}

func newTestApplierWithDir(t *testing.T) (string, *FixApplier) {
	t.Helper()

	dir := t.TempDir()

	a, err := NewFixApplier(dir)
	if err != nil {
		t.Fatalf("NewFixApplier: %v", err)
	}

	return dir, a
}

func makeFixFinding(id, before, after, file string, line int) finding.Finding {
	return finding.Finding{
		ID:          finding.ID(id),
		Rule:        "r1",
		ToolName:    "tool",
		Message:     "replace " + before + " with " + after,
		BeforeCode:  before,
		AfterCode:   after,
		Position:    finding.Position{File: finding.FilePath(file), Line: line},
		FixStrategy: finding.FixStrategyDirect,
	}
}

func makeErrorDetector(errMsg string) Detector {
	return DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
		return nil, errors.New(errMsg)
	})
}

func makeFindingDetectorFunc(id string) DetectorFunc {
	return DetectorFunc(func(_ context.Context) ([]finding.Finding, error) {
		return []finding.Finding{{ID: finding.ID(id)}}, nil
	})
}

func assertFindingErrorIO(t *testing.T, fe *finding.FindingError, file string) {
	t.Helper()

	if fe.Category != finding.ErrCategoryIO {
		t.Errorf("category = %q, want %q", fe.Category, finding.ErrCategoryIO)
	}

	if string(fe.Position.File) != file {
		t.Errorf("file = %q, want %q", fe.Position.File, file)
	}
}

func makeFixFindingWithRange(id, before, after, file string, line, col int) finding.Finding {
	return finding.Finding{
		ID:          finding.ID(id),
		BeforeCode:  before,
		AfterCode:   after,
		Position:    finding.Position{File: finding.FilePath(file), Line: line, Column: col},
		Range:       finding.NewRangePtr(finding.FilePath(file), line, col, line, col+len(before)),
		FixStrategy: finding.FixStrategyDirect,
	}
}

func makeConflictingFixes() []finding.Finding {
	return []finding.Finding{
		makeFixFindingWithRange("fix1", "old()", "new()", "fixme.go", 4, 2),
		makeFixFindingWithRange("fix2", "old()", "other()", "fixme.go", 4, 2),
	}
}

func suppressedIDs() []string { return []string{"s2", "s4"} }

func assertNoSuppressedFindings(t *testing.T, findings []finding.Finding, mode string) {
	t.Helper()

	for _, f := range findings {
		for _, sid := range suppressedIDs() {
			if string(f.ID) == sid {
				t.Errorf("suppressed finding %s present in %s results", f.ID, mode)
			}
		}
	}
}

func assertNoSuppressedNotified(t *testing.T, notified []string, mode string) {
	t.Helper()

	for _, id := range notified {
		for _, sid := range suppressedIDs() {
			if id == sid {
				t.Errorf("OnFinding called for suppressed finding %s in %s mode", id, mode)
			}
		}
	}
}

// refusingProvider CanHandles every finding but never produces edits, forcing
// the engine to fall through to later providers in the chain.
type refusingProvider struct{}

func (*refusingProvider) Name() string                   { return "refusing" }
func (*refusingProvider) CanHandle(finding.Finding) bool { return true }

func (*refusingProvider) Edits([]byte, finding.Finding) ([]FixEdit, error) {
	return nil, nil
}

// editAtBeforeProvider emulates a precise domain provider: it replaces the
// first occurrence of BeforeCode with AfterCode.
type editAtBeforeProvider struct{}

func (*editAtBeforeProvider) Name() string                     { return "edit-at-before" }
func (*editAtBeforeProvider) CanHandle(f finding.Finding) bool { return f.HasCodeChange() }

func (*editAtBeforeProvider) Edits(content []byte, f finding.Finding) ([]FixEdit, error) {
	idx := bytes.Index(content, []byte(f.BeforeCode))
	if idx < 0 {
		return nil, nil
	}

	return []FixEdit{newReplacementEdit(idx, len(f.BeforeCode), f)}, nil
}

// erroringProvider fails resolution for findings whose rule matches.
type erroringProvider struct{}

func (erroringProvider) Name() string { return "erroring" }

func (erroringProvider) CanHandle(f finding.Finding) bool {
	return f.Rule == "unresolvable"
}

func (erroringProvider) Edits(_ []byte, _ finding.Finding) ([]FixEdit, error) {
	return nil, errors.New("cannot resolve this finding")
}

// staticEditProvider always returns the given edits.
type staticEditProvider struct {
	edits []FixEdit
}

func (staticEditProvider) Name() string { return "static" }

func (staticEditProvider) CanHandle(finding.Finding) bool { return true }

func (s staticEditProvider) Edits(_ []byte, _ finding.Finding) ([]FixEdit, error) {
	return s.edits, nil
}

// saboteurProvider applies real edits while deleting the file's backup,
// simulating a backup that has become unavailable between backup and restore.
type saboteurProvider struct {
	backup     *FileBackup
	targetPath string
}

func (*saboteurProvider) Name() string { return "saboteur" }

func (*saboteurProvider) CanHandle(f finding.Finding) bool {
	return f.HasCodeChange()
}

func (s *saboteurProvider) Edits(content []byte, f finding.Finding) ([]FixEdit, error) {
	bakPath := s.backup.BackupPath(s.targetPath)
	if bakPath != "" {
		_ = os.Remove(
			bakPath,
		) //nolint:gosec // G703: intentional path manipulation in test saboteur
	}

	idx := bytes.Index(content, []byte(f.BeforeCode))
	if idx < 0 {
		return nil, errors.New("before code not found")
	}

	return []FixEdit{newReplacementEdit(idx, len(f.BeforeCode), f)}, nil
}

// cancelingProvider cancels the run's context when it sees a specific
// BeforeCode, and produces valid edits otherwise. It lets tests place the
// cancellation exactly between two files of a multi-file run.
type cancelingProvider struct {
	cancel context.CancelFunc
	before string
}

func (*cancelingProvider) Name() string                       { return "canceling" }
func (p *cancelingProvider) CanHandle(f finding.Finding) bool { return f.HasCodeChange() }
func (p *cancelingProvider) Edits(content []byte, f finding.Finding) ([]FixEdit, error) {
	if f.BeforeCode == p.before {
		p.cancel()

		return nil, errors.New("cancelled during edit resolution")
	}

	idx := bytes.Index(content, []byte(f.BeforeCode))
	if idx < 0 {
		return nil, nil
	}

	return []FixEdit{newReplacementEdit(idx, len(f.BeforeCode), f)}, nil
}

// upperProvider is a test provider that only handles direct fixes.
type upperProvider struct{}

func (upperProvider) Name() string { return "test-upper" }

func (upperProvider) CanHandle(f finding.Finding) bool {
	return f.FixStrategy == finding.FixStrategyDirect && f.BeforeCode != ""
}

func (upperProvider) Edits(_ []byte, f finding.Finding) ([]FixEdit, error) {
	return []FixEdit{
		{
			Offset:      0,
			Length:      len(f.BeforeCode),
			Replacement: []byte(f.AfterCode),
			Source:      f,
		},
	}, nil
}
