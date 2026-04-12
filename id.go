package finding

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"
)

// ID format constants
const (
	IDPartCount       = 3  // Minimum number of parts for hash-based IDs
	HashLength        = 16 // Length of hex-encoded hash
	IDPartMin         = 4  // Minimum parts to attempt parsing column and line
	MaxShortHashLength = 16 // Maximum length to identify as short hash (without :line:col)
)

// GenerateID creates a stable, unique identifier for a finding.
// Format: "tool:rule:file:line:col" (human-readable)
// If line is 0, uses hash-based ID for stability.
func GenerateID(toolName, rule string, pos Position) string {
	if pos.Line == 0 {
		// Hash-based for position-less findings
		h := sha256.New()
		h.Write([]byte(toolName + ":" + rule + ":" + pos.File))

		return fmt.Sprintf("%s:%s:%x", toolName, rule, h.Sum(nil)[:8])
	}

	// Normalize file path to use forward slashes
	file := filepath.ToSlash(pos.File)

	if pos.Column == 0 {
		return fmt.Sprintf("%s:%s:%s:%d", toolName, rule, file, pos.Line)
	}

	return fmt.Sprintf("%s:%s:%s:%d:%d", toolName, rule, file, pos.Line, pos.Column)
}

// ParseID parses a finding ID and extracts its components.
// Returns tool, rule, file, line, column, and ok status.
func ParseID(id string) (tool, rule, file string, line, column int, ok bool) {
	parts := strings.Split(id, ":")
	if len(parts) < IDPartCount {
		return "", "", "", 0, 0, false
	}

	tool = parts[0]
	rule = parts[1]

	// Handle hash-based IDs
	if len(parts) == IDPartCount && len(parts[2]) == HashLength { // hex encoded hash
		file = ""

		return tool, rule, file, 0, 0, true
	}

	// Try to parse position from remaining parts
	// Format: tool:rule:file:line or tool:rule:file:line:col
	// File may contain colons (e.g., Windows paths), so we need to be careful
	// We assume the last 1-2 parts are line:column

	if len(parts) >= IDPartMin {
		// Try parsing last part as column
		if n, err := fmt.Sscanf(parts[len(parts)-1], "%d", &column); err == nil && n == 1 {
			// Try parsing second-to-last as line
			if n2, err2 := fmt.Sscanf(parts[len(parts)-2], "%d", &line); err2 == nil && n2 == 1 {
				// File is everything between rule and line
				file = strings.Join(parts[2:len(parts)-2], ":")

				return tool, rule, file, line, column, true
			}
		}

		// No column, try line only
		if n, err := fmt.Sscanf(parts[len(parts)-1], "%d", &line); err == nil && n == 1 {
			file = strings.Join(parts[2:len(parts)-1], ":")

			return tool, rule, file, line, 0, true
		}
	}

	// Just file, no position
	file = strings.Join(parts[2:], ":")

	return tool, rule, file, 0, 0, true
}

// IsHashID returns true if the ID appears to be hash-based.
func IsHashID(id string) bool {
	parts := strings.Split(id, ":")
	if len(parts) != IDPartCount {
		return false
	}

	return len(parts[2]) == HashLength // 8 bytes hex encoded
}
