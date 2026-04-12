package finding

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"
)

// ID format constants.
const (
	IDPartCount        = 3  // Minimum number of parts for hash-based IDs
	HashLength         = 16 // Length of hex-encoded hash
	IDPartMin          = 4  // Minimum parts to attempt parsing column and line
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

// extractFile extracts the file path from ID parts, excluding trailing position components.
// The parts slice is expected to be [tool, rule, file parts..., line?, column?].
// trailingCount is the number of trailing position parts (1 for line only, 2 for line:col).
func extractFile(parts []string, trailingCount int) string {
	if len(parts) < IDPartCount+1 { // Need at least tool:rule:file (3 parts)
		return ""
	}

	return strings.Join(parts[2:len(parts)-trailingCount], ":")
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
		return tool, rule, "", 0, 0, true
	}

	// Try to parse position from remaining parts
	// Format: tool:rule:file:line or tool:rule:file:line:col
	// File may contain colons (e.g., Windows paths), so we need to be careful
	// We assume the last 1-2 parts are line:column

	if len(parts) >= IDPartMin {
		// Try parsing last part as column
		err := parseInt(parts[len(parts)-1], &column)
		if err == nil {
			// Try parsing second-to-last as line
			err2 := parseInt(parts[len(parts)-2], &line)
			if err2 == nil {
				file = extractFile(parts, 2)

				return tool, rule, file, line, column, true
			}
		}

		// No column, try line only
		parseErr := parseInt(parts[len(parts)-1], &line)
		if parseErr == nil {
			file = extractFile(parts, 1)

			return tool, rule, file, line, 0, true
		}
	}

	// Just file, no position
	file = strings.Join(parts[2:], ":")

	return tool, rule, file, 0, 0, true
}

// parseInt is a helper to parse a string to int, returning nil on success.
func parseInt(s string, result *int) error {
	_, err := fmt.Sscanf(s, "%d", result)
	if err != nil {
		return fmt.Errorf("failed to parse int: %w", err)
	}

	return nil
}

// IsHashID returns true if the ID appears to be hash-based.
func IsHashID(id string) bool {
	parts := strings.Split(id, ":")
	if len(parts) != IDPartCount {
		return false
	}

	return len(parts[2]) == HashLength // 8 bytes hex encoded
}
