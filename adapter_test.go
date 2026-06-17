package finding

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

type testOutput struct {
	Issues []testIssue `json:"issues"`
}

type testIssue struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
	File    string `json:"file"`
	Line    int    `json:"line"`
}

func testParse(data []byte) (testOutput, error) {
	var out testOutput

	err := json.Unmarshal(data, &out)
	if err != nil {
		return testOutput{}, err
	}

	return out, nil
}

func testConvert(out testOutput) ([]Finding, error) {
	findings := make([]Finding, 0, len(out.Issues))
	for _, issue := range out.Issues {
		findings = append(findings, Finding{
			ID:       GenerateID("test", issue.Rule, Position{File: issue.File, Line: issue.Line}),
			Rule:     issue.Rule,
			ToolName: "test-tool",
			Message:  issue.Message,
			Severity: SeverityError,
			Position: Position{File: issue.File, Line: issue.Line},
		})
	}

	return findings, nil
}

func TestToolAdapter_Detect(t *testing.T) {
	t.Parallel()

	data := testOutput{Issues: []testIssue{
		{Rule: "no-unused", Message: "unused variable", File: "main.go", Line: 10},
		{Rule: "sql-inject", Message: "SQL injection", File: "db.go", Line: 42},
	}}

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	run := func(_ context.Context) ([]byte, error) { return raw, nil }

	adapter := NewToolAdapter("test-tool", run, testParse, testConvert)

	findings, err := adapter.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}

	if len(findings) != 2 {
		t.Fatalf("got %d findings, want 2", len(findings))
	}

	if findings[0].Rule != "no-unused" {
		t.Errorf("findings[0].Rule = %q, want %q", findings[0].Rule, "no-unused")
	}

	if findings[1].Rule != "sql-inject" {
		t.Errorf("findings[1].Rule = %q, want %q", findings[1].Rule, "sql-inject")
	}
}

func TestToolAdapter_EmptyOutput(t *testing.T) {
	t.Parallel()

	run := func(_ context.Context) ([]byte, error) { return nil, nil }
	adapter := NewToolAdapter("empty-tool", run, testParse, testConvert)

	findings, err := adapter.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}

	if findings != nil {
		t.Errorf("expected nil findings for empty output, got %v", findings)
	}
}

func TestToolAdapter_RunError(t *testing.T) {
	t.Parallel()

	runErr := errors.New("tool crashed")
	run := func(_ context.Context) ([]byte, error) { return nil, runErr }
	adapter := NewToolAdapter("failing-tool", run, testParse, testConvert)

	findings, err := adapter.Detect(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if findings != nil {
		t.Errorf("expected nil findings on error, got %v", findings)
	}

	if !errors.Is(err, runErr) {
		t.Errorf("error should wrap runErr, got: %v", err)
	}
}

func TestToolAdapter_ParseError(t *testing.T) {
	t.Parallel()

	run := func(_ context.Context) ([]byte, error) { return []byte("not json"), nil }
	adapter := NewToolAdapter("bad-parse", run, testParse, testConvert)

	_, err := adapter.Detect(context.Background())
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
}

func TestToolAdapter_Name(t *testing.T) {
	t.Parallel()

	run := func(_ context.Context) ([]byte, error) { return nil, nil }
	adapter := NewToolAdapter("my-tool", run, testParse, testConvert)

	if adapter.Name() != "my-tool" {
		t.Errorf("Name() = %q, want %q", adapter.Name(), "my-tool")
	}
}

func TestToolAdapter_ImplementsDetector(t *testing.T) {
	t.Parallel()

	run := func(_ context.Context) ([]byte, error) { return nil, nil }
	adapter := NewToolAdapter("iface-test", run, testParse, testConvert)

	var _ Detector = adapter
}

func TestToolAdapter_NilPanics(t *testing.T) {
	t.Parallel()

	run := func(_ context.Context) ([]byte, error) { return nil, nil }

	t.Run("nil run", func(t *testing.T) {
		t.Parallel()

		AssertPanics(t, "expected panic for nil function", func() {
			NewToolAdapter[testOutput]("x", nil, testParse, testConvert)
		})
	})

	t.Run("nil parse", func(t *testing.T) {
		t.Parallel()

		AssertPanics(t, "expected panic for nil function", func() {
			NewToolAdapter[testOutput]("x", run, nil, testConvert)
		})
	})

	t.Run("nil convert", func(t *testing.T) {
		t.Parallel()

		AssertPanics(t, "expected panic for nil function", func() {
			NewToolAdapter[testOutput]("x", run, testParse, nil)
		})
	})
}
