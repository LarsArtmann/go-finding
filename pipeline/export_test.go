package pipeline

import "github.com/larsartmann/go-finding"

// rangeFix is the unexported canonical test helper. It is exposed to the
// external test package (pipeline_test) via ExportRangeFix below.
//
// Standard Go pattern: export_test.go bridges unexported test helpers to the
// external test package without polluting the public API of the pipeline
// package.
func rangeFix(
	file string,
	startLine, startCol, endLine, endCol int,
	before, after string,
) finding.Finding {
	return finding.Finding{
		BeforeCode: before,
		AfterCode:  after,
		Range:      finding.NewRangePtr(finding.FilePath(file), startLine, startCol, endLine, endCol),
		Position:   finding.Pos(finding.FilePath(file), startLine, startCol),
	}
}

// offsetFix is the unexported canonical test helper for byte-offset ranges.
func offsetFix(before, after string, startOff, endOff int) finding.Finding {
	return finding.Finding{
		BeforeCode: before,
		AfterCode:  after,
		Range: &finding.Range{
			Start: finding.Position{File: "a.go", Offset: startOff},
			End:   finding.Position{File: "a.go", Offset: endOff},
		},
		Position: finding.Pos("a.go", 4, 2),
	}
}

// offsetFixWithID is offsetFix with an explicit Finding.ID.
func offsetFixWithID(id, before, after string, startOff, endOff int) finding.Finding {
	f := offsetFix(before, after, startOff, endOff)
	f.ID = finding.ID(id)

	return f
}

// ExportRangeFix exposes rangeFix to the external test package (pipeline_test).
func ExportRangeFix(
	file string,
	startLine, startCol, endLine, endCol int,
	before, after string,
) finding.Finding {
	return rangeFix(file, startLine, startCol, endLine, endCol, before, after)
}

// ExportOffsetFix exposes offsetFix to the external test package (pipeline_test).
func ExportOffsetFix(before, after string, startOff, endOff int) finding.Finding {
	return offsetFix(before, after, startOff, endOff)
}

// ExportOffsetFixWithID exposes offsetFixWithID to the external test package
// (pipeline_test).
func ExportOffsetFixWithID(id, before, after string, startOff, endOff int) finding.Finding {
	return offsetFixWithID(id, before, after, startOff, endOff)
}
