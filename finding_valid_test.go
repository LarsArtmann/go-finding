package finding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFinding_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		f    Finding
		want bool
	}{
		{"valid", NewFinding("r", "t", "m", SeverityError, Pos("a.go", 1, 1)), true},
		{"empty id", Finding{Rule: "r", ToolName: "t", Message: "m", Severity: SeverityError, Position: Pos("a.go", 1, 1)}, false},
		{"empty rule", Finding{ID: "1", ToolName: "t", Message: "m", Severity: SeverityError, Position: Pos("a.go", 1, 1)}, false},
		{"empty tool", Finding{ID: "1", Rule: "r", Message: "m", Severity: SeverityError, Position: Pos("a.go", 1, 1)}, false},
		{"empty message", Finding{ID: "1", Rule: "r", ToolName: "t", Severity: SeverityError, Position: Pos("a.go", 1, 1)}, false},
		{"invalid position", Finding{ID: "1", Rule: "r", ToolName: "t", Message: "m", Severity: SeverityError, Position: Position{File: "a.go", Line: -1}}, false},
		{"invalid severity", Finding{ID: "1", Rule: "r", ToolName: "t", Message: "m", Severity: "bogus", Position: Pos("a.go", 1, 1)}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.f.IsValid())
		})
	}
}
