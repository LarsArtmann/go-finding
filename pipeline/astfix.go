// Package pipeline provides a detect → triage → fix → verify workflow
// for automated code remediation.
package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"

	"github.com/larsartmann/go-finding"
)

// ASTFixer applies fixes using AST transformations when possible,
// falling back to text replacement for non-parsable code.
type ASTFixer struct {
	fset *token.FileSet
}

// NewASTFixer creates a new AST-based fix applier.
func NewASTFixer() *ASTFixer {
	return &ASTFixer{
		fset: token.NewFileSet(),
	}
}

// ApplyResult represents the outcome of applying a fix.
type ApplyResult struct {
	Applied     bool
	Method      string // "ast" or "text"
	Error       error
	BackupPath  string
}

// ApplyFixes applies multiple fixes to a file with conflict detection.
func (a *ASTFixer) ApplyFixes(ctx context.Context, filePath string, fixes []finding.Finding) ([]ApplyResult, error) {
	// Detect conflicts first
	detector := NewConflictDetector()
	groups, conflicts := detector.DetectConflicts(fixes)

	results := make([]ApplyResult, 0, len(fixes))

	// Apply non-conflicting fixes
	for _, group := range groups {
		for _, f := range group.Fixes {
			result := a.Apply(ctx, filePath, f)
			results = append(results, result)
		}
	}

	// Mark conflicting fixes as skipped
	for range conflicts {
		results = append(results, ApplyResult{
			Applied: false,
			Method:  "skipped",
			Error:   fmt.Errorf("conflicts with another fix"),
		})
	}

	return results, nil
}

// Apply attempts to apply a single fix to a file.
// It tries AST-based transformation first, then falls back to text replacement.
func (a *ASTFixer) Apply(ctx context.Context, filePath string, fix finding.Finding) ApplyResult {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ApplyResult{
			Applied: false,
			Error:   fmt.Errorf("read file: %w", err),
		}
	}

	// Try AST-based fix first
	if result, ok := a.tryASTFix(ctx, filePath, content, fix); ok {
		return result
	}

	// Fall back to text replacement
	return a.applyTextFix(filePath, content, fix)
}

// tryASTFix attempts to apply a fix using AST transformations.
// Returns (result, true) if successful, (_, false) if should fall back.
func (a *ASTFixer) tryASTFix(ctx context.Context, filePath string, content []byte, fix finding.Finding) (ApplyResult, bool) {
	// Parse the file
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		// Can't parse - fall back to text replacement
		return ApplyResult{}, false
	}

	// Check if we can apply an AST transformation for this fix
	if fix.BeforeCode == "" || fix.AfterCode == "" {
		// No code replacement specified - fall back
		return ApplyResult{}, false
	}

	// Try to find and replace the node
	modified := a.findAndReplaceNode(f, fset, fix.BeforeCode, fix.AfterCode)
	if !modified {
		// Couldn't find the node - fall back to text replacement
		return ApplyResult{}, false
	}

	// Format the modified AST
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, f); err != nil {
		return ApplyResult{
			Applied: false,
			Error:   fmt.Errorf("format modified AST: %w", err),
		}, true
	}

	// Write the result
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return ApplyResult{
			Applied: false,
			Error:   fmt.Errorf("write file: %w", err),
		}, true
	}

	return ApplyResult{
		Applied: true,
		Method:  "ast",
	}, true
}

// findAndReplaceNode attempts to find a node matching beforeCode and replace it with afterCode.
// This is a simplified implementation that handles common cases.
// Returns true if a modification was made.
func (a *ASTFixer) findAndReplaceNode(f *ast.File, fset *token.FileSet, beforeCode, afterCode string) bool {
	// For now, this is a placeholder for AST-based transformations
	// A full implementation would:
	// 1. Walk the AST looking for nodes matching beforeCode
	// 2. Replace those nodes with the parsed afterCode
	// 3. Handle edge cases (comments, formatting, etc.)

	// Since implementing full AST matching is complex, we'll signal
	// that we couldn't apply an AST fix and let it fall back to text replacement
	return false
}

// applyTextFix applies a fix using simple text replacement.
func (a *ASTFixer) applyTextFix(filePath string, content []byte, fix finding.Finding) ApplyResult {
	if fix.BeforeCode == "" || fix.AfterCode == "" {
		return ApplyResult{
			Applied: false,
			Method:  "text",
			Error:   fmt.Errorf("no before/after code specified"),
		}
	}

	result := string(content)
	result = replaceAllString(result, fix.BeforeCode, fix.AfterCode)

	// Check if replacement actually happened
	if result == string(content) {
		return ApplyResult{
			Applied: false,
			Method:  "text",
			Error:   fmt.Errorf("code to replace not found"),
		}
	}

	if err := os.WriteFile(filePath, []byte(result), 0644); err != nil {
		return ApplyResult{
			Applied: false,
			Method:  "text",
			Error:   fmt.Errorf("write file: %w", err),
		}
	}

	return ApplyResult{
		Applied: true,
		Method:  "text",
	}
}

// VerifySyntax verifies that the file at the given path has valid Go syntax.
func (a *ASTFixer) VerifySyntax(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, filePath, content, parser.AllErrors)

	return err
}

// CanApplyASTFix checks if a fix can potentially be applied using AST transformations.
func (a *ASTFixer) CanApplyASTFix(filePath string, fix finding.Finding) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return false
	}

	// For now, only support fixes with explicit before/after code
	return fix.BeforeCode != "" && fix.AfterCode != ""
}

// ASTApplyStats tracks statistics about AST vs text-based fix application.
type ASTApplyStats struct {
	ASTApplied   int
	TextApplied  int
	Failed       int
	Skipped      int
}

// Total returns the total number of fixes processed.
func (s ASTApplyStats) Total() int {
	return s.ASTApplied + s.TextApplied + s.Failed + s.Skipped
}

// CalculateStats calculates statistics from a slice of apply results.
func CalculateStats(results []ApplyResult) ASTApplyStats {
	var stats ASTApplyStats
	for _, r := range results {
		switch {
		case !r.Applied && r.Method == "skipped":
			stats.Skipped++
		case !r.Applied:
			stats.Failed++
		case r.Method == "ast":
			stats.ASTApplied++
		case r.Method == "text":
			stats.TextApplied++
		}
	}
	return stats
}
