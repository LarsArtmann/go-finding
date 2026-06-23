package finding

// Branded primitive types for compile-time safety.
// These prevent accidental mixing of string fields that represent
// distinct domain concepts (e.g., putting a Rule where an ID belongs).
// JSON serialization is identical to raw string — branded types marshal
// as strings with no overhead.

// ID is the stable unique identifier for a finding (tool:rule:file:line:col).
type ID string

// RuleName is the rule or check name (e.g., "nilcheck", "STRONG_ID").
type RuleName string

// ToolName is the source tool name (e.g., "govet", "staticcheck").
type ToolName string

// FilePath is a path to a source file.
type FilePath string
