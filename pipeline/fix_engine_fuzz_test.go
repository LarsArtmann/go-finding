package pipeline

import (
	"bytes"
	"cmp"
	"slices"
	"testing"
)

// FuzzApplyEditsToContent verifies that the optimized single-pass edit application
// produces identical results to a naive O(n²) one-at-a-time approach.
// Edits are generated as sequential, non-overlapping operations.
func FuzzApplyEditsToContent(f *testing.F) {
	// Seed cases: (content, spec)
	f.Add([]byte("hello"), []byte{5, 0})                // delete entire content
	f.Add([]byte("hello world"), []byte{5, 3})          // replace first 5 bytes with XXX
	f.Add([]byte("hello"), []byte{})                    // no edits
	f.Add([]byte("ab"), []byte{1, 0, 1, 0})             // two adjacent deletions
	f.Add([]byte("test"), []byte{0, 4, 0})              // delete all, then nothing
	f.Add([]byte("abcdefgh"), []byte{2, 1, 2, 1, 2, 1}) // three small replacements
	f.Add([]byte("x"), []byte{0, 1, 5})                 // replace single char

	f.Fuzz(func(t *testing.T, content, spec []byte) {
		if len(content) > 1000 || len(spec) > 300 {
			return
		}

		edits := sequentialEdits(content, spec)
		if len(edits) == 0 {
			return
		}

		// Sort descending by offset (required by applyEditsToContent)
		slices.SortFunc(edits, func(a, b FixEdit) int {
			return cmp.Compare(b.Offset, a.Offset)
		})

		result := applyEditsToContent(content, edits)

		// Property 1: result is never nil
		if result == nil {
			t.Fatal("applyEditsToContent returned nil")
		}

		// Property 2: result matches naive O(n²) application
		naive := naiveApplyEdits(content, edits)
		if !bytes.Equal(result, naive) {
			t.Errorf(
				"result mismatch for content=%q edits=%v:\nfast=%q\nnaive=%q",
				content, edits, result, naive,
			)
		}

		// Property 3: expected length is deterministic
		expectedLen := len(content)
		for _, e := range edits {
			expectedLen += len(e.Replacement) - e.Length
		}

		if len(result) != expectedLen {
			t.Errorf("length mismatch: got %d, want %d (content=%d)", len(result), expectedLen, len(content))
		}
	})
}

// FuzzApplyEditsToContent_Overlapping verifies that applyEditsToContent does not
// panic even when edits overlap (it is not required to produce correct results
// for overlapping edits, but it must not crash).
func FuzzApplyOverlappingEdits(f *testing.F) {
	f.Add([]byte("hello"), []byte{0, 3, 1, 3})
	f.Add([]byte("test"), []byte{0, 2, 1, 2, 2, 1})

	f.Fuzz(func(t *testing.T, content, spec []byte) {
		_ = t

		if len(content) > 500 || len(spec) > 200 {
			return
		}

		edits := overlappingEdits(content, spec)
		if len(edits) == 0 {
			return
		}

		slices.SortFunc(edits, func(a, b FixEdit) int {
			return cmp.Compare(b.Offset, a.Offset)
		})

		// Must not panic
		_ = applyEditsToContent(content, edits)
	})
}

// sequentialEdits generates non-overlapping edits from the fuzz spec.
// Each pair of bytes encodes (length, replacementLength). Edits are applied
// sequentially starting from offset 0.
func sequentialEdits(content, spec []byte) []FixEdit {
	var edits []FixEdit

	pos := 0

	for i := 0; i+1 < len(spec) && pos <= len(content); i += 2 {
		maxLen := len(content) - pos
		if maxLen <= 0 {
			break
		}

		length := int(spec[i])%maxLen + 1 // 1..maxLen, never 0 (avoids same-offset overlaps)
		repLen := int(spec[i+1]) % 8

		edits = append(edits, FixEdit{
			Offset:      pos,
			Length:      length,
			Replacement: bytes.Repeat([]byte{'X'}, repLen),
		})

		pos += length
	}

	return edits
}

// overlappingEdits generates potentially overlapping edits from the fuzz spec.
func overlappingEdits(content, spec []byte) []FixEdit {
	if len(content) == 0 {
		return nil
	}

	var edits []FixEdit

	for i := 0; i+1 < len(spec); i += 2 {
		offset := int(spec[i]) % len(content)

		maxLen := len(content) - offset
		if maxLen <= 0 {
			continue
		}

		length := int(spec[i+1]) % (maxLen + 1)
		repLen := (int(spec[i]) + int(spec[i+1])) % 5

		edit := FixEdit{
			Offset:      offset,
			Length:      length,
			Replacement: bytes.Repeat([]byte{'Z'}, repLen),
		}

		edits = append(edits, edit)
	}

	return edits
}

// naiveApplyEdits applies edits one at a time (O(n²)). Edits must be sorted
// descending by offset. Used as the oracle for fuzz testing.
func naiveApplyEdits(content []byte, edits []FixEdit) []byte {
	result := make([]byte, len(content))
	copy(result, content)

	for _, edit := range edits {
		if edit.Offset < 0 || edit.Offset+edit.Length > len(result) {
			continue
		}

		tmp := make([]byte, 0, len(result)-edit.Length+len(edit.Replacement))
		tmp = append(tmp, result[:edit.Offset]...)
		tmp = append(tmp, edit.Replacement...)
		tmp = append(tmp, result[edit.Offset+edit.Length:]...)
		result = tmp
	}

	return result
}
