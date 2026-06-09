package pipeline

import (
	"bytes"
	"log/slog"
	"testing"
	"testing/fstest"

	"github.com/LarsArtmann/gogenfilter/v3"
	"github.com/larsartmann/go-finding"
)

func TestGeneratedFileFilter_LogFilterError(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"main.go": {Data: []byte("package main\n")},
	}

	var buf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&buf, nil))

	opt, err := gogenfilter.WithFilterOptions(gogenfilter.FilterAll)
	if err != nil {
		t.Fatalf("WithFilterOptions() error = %v", err)
	}

	gf, err := NewGeneratedFileFilter(
		logger,
		opt,
		gogenfilter.WithFS(fsys),
	)
	if err != nil {
		t.Fatalf("NewGeneratedFileFilter() error = %v", err)
	}

	findings := []finding.Finding{
		{ID: "1", Position: finding.Position{File: "missing.go"}},
	}

	result, err := gf.Process(t.Context(), findings)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Process() = %d findings, want 1 (fail-open)", len(result))
	}

	logOutput := buf.String()
	if logOutput == "" {
		t.Fatal("expected log output for missing file error, got empty")
	}
}

func TestGeneratedFileFilter_LogFilterError_NilLogger(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{}

	opt, err := gogenfilter.WithFilterOptions(gogenfilter.FilterAll)
	if err != nil {
		t.Fatalf("WithFilterOptions() error = %v", err)
	}

	gf, err := NewGeneratedFileFilter(
		nil,
		opt,
		gogenfilter.WithFS(fsys),
	)
	if err != nil {
		t.Fatalf("NewGeneratedFileFilter() error = %v", err)
	}

	findings := []finding.Finding{
		{ID: "1", Position: finding.Position{File: "missing.go"}},
	}

	result, err := gf.Process(t.Context(), findings)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Process() = %d findings, want 1 (fail-open with nil logger)", len(result))
	}
}
