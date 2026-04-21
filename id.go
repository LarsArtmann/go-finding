package finding

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// ID format constants.
const (
	idPartCount      = 3  // Minimum number of parts for hash-based IDs
	hashLength       = 16 // Length of hex-encoded hash
	idPartMin        = 4  // Minimum parts to attempt parsing column and line
	positionPartsOne = 1  // Number of positional parts when only line is present
	positionPartsTwo = 2  // Number of positional parts when line and column are present
)

// GenerateID creates a stable, unique identifier for a finding.
// Format: "tool:rule:file:line:col" (human-readable)
// If line is 0, uses hash-based ID for stability.
func GenerateID(toolName, rule string, pos Position) string {
	if pos.Line == 0 {
		// Hash-based for position-less findings
		h := sha256.New()
		h.Write([]byte(toolName + ":" + rule + ":" + pos.File))

		return fmt.Sprintf("%s:%s:%x", toolName, rule, h.Sum(nil)[:hashLength/2])
	}

	// Normalize file path to use forward slashes
	file := filepath.ToSlash(pos.File)

	if pos.Column == 0 {
		return fmt.Sprintf("%s:%s:%s:%d", toolName, rule, file, pos.Line)
	}

	return fmt.Sprintf("%s:%s:%s:%d:%d", toolName, rule, file, pos.Line, pos.Column)
}

// extractFile extracts the file path from ID parts, excluding trailing position components.
// The parts slice is expected to be [tool, rule, file parts..., line?, column?].
// trailingCount is the number of trailing position parts (1 for line only, 2 for line:col).
func extractFile(parts []string, trailingCount int) string {
	if len(parts) < idPartCount+1 { // Need at least tool:rule:file (3 parts)
		return ""
	}

	return strings.Join(parts[2:len(parts)-trailingCount], ":")
}

// ParsedID holds the components of a parsed finding ID.
type ParsedID struct {
	Tool   string
	Rule   string
	File   string
	Line   int
	Column int
}

// OK returns true if the ID was successfully parsed.
func (p ParsedID) OK() bool {
	return p.Tool != ""
}

// ParseID parses a finding ID and extracts its components.
func ParseID(id string) ParsedID {
	parts := strings.Split(id, ":")

	if len(parts) < idPartCount {
		return ParsedID{} //nolint:exhaustruct
	}

	tool := parts[0]
	rule := parts[1]

	// Handle hash-based IDs
	if len(parts) == idPartCount && len(parts[2]) == hashLength { // hex encoded hash
		return ParsedID{Tool: tool, Rule: rule} //nolint:exhaustruct
	}

	// Try to parse position from remaining parts
	// Format: tool:rule:file:line or tool:rule:file:line:col
	// File may contain colons (e.g., Windows paths), so we need to be careful
	// We assume the last 1-2 parts are line:column

	if len(parts) >= idPartMin {
		// Try parsing last part as column
		colTest := 0

		err := parseInt(parts[len(parts)-1], &colTest)
		if err == nil {
			// Try parsing second-to-last as line
			lineTest := 0

			err2 := parseInt(parts[len(parts)-2], &lineTest)
			if err2 == nil {
				file := extractFile(parts, positionPartsTwo)

				return ParsedID{Tool: tool, Rule: rule, File: file, Line: lineTest, Column: colTest}
			}
		}

		// No column, try line only
		var line int

		err = parseInt(parts[len(parts)-1], &line)
		if err == nil {
			file := extractFile(parts, positionPartsOne)

			return ParsedID{Tool: tool, Rule: rule, File: file, Line: line} //nolint:exhaustruct
		}
	}

	// Just file, no position
	file := strings.Join(parts[2:], ":")

	return ParsedID{Tool: tool, Rule: rule, File: file} //nolint:exhaustruct
}

// parseInt is a helper to parse a string to int, returning nil on success.
func parseInt(s string, result *int) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("failed to parse %q as int: %w", s, err)
	}

	*result = n

	return nil
}

// IsHashID returns true if the ID appears to be hash-based.
func IsHashID(id string) bool {
	parts := strings.Split(id, ":")
	if len(parts) != idPartCount {
		return false
	}

	return isHexString(parts[2]) && len(parts[2]) == hashLength
}

// isHexString reports whether s consists entirely of hex digits.
func isHexString(s string) bool {
	for _, r := range s {
		if !isHexDigit(r) {
			return false
		}
	}

	return len(s) > 0
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}
