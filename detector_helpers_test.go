package finding

import (
	"context"
	"errors"
	"testing"
)

func TestCheckBinary_NotFound(t *testing.T) {
	t.Parallel()

	_, err := CheckBinary("this-binary-definitely-does-not-exist-12345")
	if err == nil {
		t.Fatal("expected error for nonexistent binary")
	}

	if !errors.Is(err, ErrIO) {
		t.Errorf("error = %v, want ErrIO", err)
	}
}

func TestCheckBinary_Found(t *testing.T) {
	t.Parallel()

	// "true" is a built-in on all Unix systems
	path, err := CheckBinary("true")
	if err != nil {
		t.Fatalf("CheckBinary(\"true\") error: %v", err)
	}

	if path == "" {
		t.Error("expected non-empty path")
	}
}

func TestRunCmd_Success(t *testing.T) {
	t.Parallel()

	// "true" exits 0 with no output
	output, err := RunCmd(context.Background(), "true")
	if err != nil {
		t.Fatalf("RunCmd(\"true\") error: %v", err)
	}

	if len(output) != 0 {
		t.Errorf("expected empty output, got %q", string(output))
	}
}

func TestRunCmd_Failure(t *testing.T) {
	t.Parallel()

	// "false" exits 1
	_, err := RunCmd(context.Background(), "false")
	if err == nil {
		t.Fatal("expected error for failing command")
	}

	if !errors.Is(err, ErrIO) {
		t.Errorf("error = %v, want ErrIO", err)
	}
}

func TestRunCmd_Output(t *testing.T) {
	t.Parallel()

	// "echo" outputs its argument
	output, err := RunCmd(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("RunCmd error: %v", err)
	}

	if string(output) != "hello\n" {
		t.Errorf("output = %q, want %q", string(output), "hello\n")
	}
}

func TestSeverityPriorityString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		sev  Severity
		want string
	}{
		{SeverityCritical, "critical"},
		{SeverityError, "high"},
		{SeverityWarning, "medium"},
		{SeverityInfo, "low"},
		{Severity("bogus"), "bogus"},
	}

	for _, tt := range tests {
		t.Run(string(tt.sev), func(t *testing.T) {
			t.Parallel()

			got := tt.sev.PriorityString()
			if got != tt.want {
				t.Errorf("PriorityString() = %q, want %q", got, tt.want)
			}
		})
	}
}
