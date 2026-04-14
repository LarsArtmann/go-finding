package pipeline

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/larsartmann/go-finding"
)

// ASTFixer applies text-based fixes to Go files with optional syntax verification.
// Despite the name, the current implementation uses text replacement rather than
// AST transformations. VerifySyntax can be used to validate Go syntax after applying fixes.
type ASTFixer struct{}

// NewASTFixer creates a new AST-based fix applier.
func NewASTFixer() *ASTFixer {
	return &ASTFixer{}
}

// ApplyResult represents the outcome of applying a fix.
type ApplyResult struct {
	Applied     bool
	Method      string // "text" or "skipped"
	Error       error
	BackupPath  string
}

// ApplyFixes applies multiple fixes to a file with conflict detection.
func (a *ASTFixer) ApplyFixes(ctx context.Context, filePath string, fixes []finding.Finding) ([]ApplyResult, error) {
	detector := NewConflictDetector()
	groups, conflicts := detector.DetectConflicts(fixes)

	results := make([]ApplyResult, 0, len(fixes))

	for _, group := range groups {
		for _, f := range group.Fixes {
			result := a.Apply(ctx, filePath, f)
			results = append(results, result)
		}
	}

	for _, cf := range conflicts {
		results = append(results, ApplyResult{
			Applied: false,
			Method:  "skipped",
			Error:   fmt.Errorf("conflicts with another fix: %s", cf.ID),
		})
	}

	return results, nil
}

// Apply attempts to apply a single fix to a file using text replacement.
func (a *ASTFixer) Apply(_ context.Context, filePath string, fix finding.Finding) ApplyResult {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ApplyResult{
			Applied: false,
			Error:   fmt.Errorf("read file: %w", err),
		}
	}

	if fix.BeforeCode == "" || fix.AfterCode == "" {
		return ApplyResult{
			Applied: false,
			Method:  "text",
			Error:   fmt.Errorf("no before/after code specified"),
		}
	}

	result := string(content)
	result = strings.ReplaceAll(result, fix.BeforeCode, fix.AfterCode)

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

// CanApplyASTFix checks if a fix can potentially be applied to a valid Go file.
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

	return fix.BeforeCode != "" && fix.AfterCode != ""
}

// ASTApplyStats tracks statistics about text-based fix application.
type ASTApplyStats struct {
	TextApplied int
	Failed      int
	Skipped     int
}

// Total returns the total number of fixes processed.
func (s ASTApplyStats) Total() int {
	return s.TextApplied + s.Failed + s.Skipped
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
		case r.Method == "text":
			stats.TextApplied++
		}
	}
	return stats
}
