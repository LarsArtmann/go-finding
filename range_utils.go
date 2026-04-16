package finding

// RangeLinesEq checks if two ranges have equal start/end lines (ignoring columns/files).
func RangeLinesEq(a, b Range) bool {
	return a.Start.Line == b.Start.Line && a.End.Line == b.End.Line
}
