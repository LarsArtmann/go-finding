package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-output"
	_ "github.com/larsartmann/go-output/delimited" // registers CSV + TSV renderers
	_ "github.com/larsartmann/go-output/markdown"  // registers Markdown renderer
)

// goOutputFormats is the set of formats delegated to go-output's RenderTable.
var goOutputFormats = map[string]output.Format{
	"markdown": output.FormatMarkdown,
	"csv":      output.FormatCSV,
	"tsv":      output.FormatTSV,
}

// isGoOutputFormat returns true if the format string is handled by go-output.
func isGoOutputFormat(format string) bool {
	_, ok := goOutputFormats[format]

	return ok
}

// findingToTableData converts a slice of Findings to go-output's Table.
// This is the adapter between go-finding's domain model and go-output's presentation layer.
func findingToTableData(findings []finding.Finding) *output.Table {
	data := output.NewTable([]string{"Location", "Severity", "Category", "Rule", "Message", "Fix"})

	for _, f := range findings {
		data.AddRow([]string{
			f.Position.String(),
			strings.ToUpper(string(f.Severity)),
			string(f.Category),
			string(f.Rule),
			f.Message,
			string(f.FixStrategy),
		})
	}

	return data
}

// renderGoOutput renders findings using go-output's format dispatch.
// Markdown gets a title with finding count; CSV/TSV are clean data exports
// with no footer or title so downstream parsers aren't broken.
func renderGoOutput(w io.Writer, report *finding.Report, format string) error {
	findings := report.FindingsSnapshot()

	if len(findings) == 0 {
		return nil
	}

	data := findingToTableData(findings)

	f, ok := goOutputFormats[format]
	if !ok {
		return fmt.Errorf("unsupported go-output format: %s", format)
	}

	opts := output.RenderOptions{
		Writer:    w,
		ColorMode: output.ColorModeAuto,
	}

	if format == "markdown" {
		opts.Title = fmt.Sprintf("Findings (%d)", len(findings))
	}

	return output.RenderTable(data, f, opts)
}
