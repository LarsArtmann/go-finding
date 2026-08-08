package pipeline

import (
	"strings"
	"testing"
)

// FuzzSanitizeFilename verifies that sanitizeFilename always produces a safe,
// non-empty filename consisting only of alphanumeric characters and single hyphens.
func FuzzSanitizeFilename(f *testing.F) {
	f.Add("detect")
	f.Add("slow-stage-30s")
	f.Add("")
	f.Add("!!!")
	f.Add("a---b")
	f.Add("/etc/passwd")
	f.Add("café-münchen")
	f.Add("stage\u0000null")
	f.Add("very-long-reason-" + strings.Repeat("x", 200))

	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 500 {
			return
		}

		result := sanitizeFilename(input)

		if result == "" {
			t.Fatal("sanitizeFilename returned empty string")
		}

		if strings.Contains(result, "--") {
			t.Fatalf("result contains consecutive hyphens: %q", result)
		}

		if strings.HasPrefix(result, "-") || strings.HasSuffix(result, "-") {
			t.Fatalf("result has leading/trailing hyphen: %q", result)
		}

		for _, r := range result {
			isAlphaNum := (r >= 'a' && r <= 'z') ||
				(r >= 'A' && r <= 'Z') ||
				(r >= '0' && r <= '9') ||
				r == '-'
			if !isAlphaNum {
				t.Fatalf("result contains non-safe character %q in %q", r, result)
			}
		}
	})
}
