package pipeline

import (
	"errors"
	"testing"

	"github.com/larsartmann/go-finding"
)

// FuzzEditListProvider exercises the EditListProvider resolution path
// (direct offsets and line/column via the shared line index) plus the
// single-pass applier, against arbitrary content and coordinates.
//
// Invariants: never panics; an error result carries no edits; resolved
// edits are in bounds; application of resolved edits never panics.
func FuzzEditListProvider(f *testing.F) {
	f.Add([]byte("abc\ndef\nghi\n"), 0, 1, 1, 3, 1, 4, "x")   // byte span line 1
	f.Add([]byte("old()\nold()\n"), 5, 2, 1, 10, 2, 6, "n()") // line/col span line 2
	f.Add([]byte("insertion target"), 7, 0, 0, -1, 0, 0, "<ins>")
	f.Add([]byte(""), 0, 0, 0, 0, 0, 0, "")
	f.Add([]byte("tiny"), 900, 99, 99, 1000, 99, 99, "out of range")
	f.Add([]byte("negative"), -1, 0, 0, 3, 0, 0, "neg")

	f.Fuzz(func(t *testing.T, content []byte, so, sl, sc, eo, el, ec int, text string) {
		edit := finding.TextEdit{
			Start:   finding.Position{File: "f.go", Offset: so, Line: sl, Column: sc},
			End:     finding.Position{Offset: eo, Line: el, Column: ec},
			NewText: text,
		}

		target := finding.Finding{
			Position:    finding.Position{File: "f.go"},
			FixStrategy: finding.FixStrategyDirect,
			Edits:       []finding.TextEdit{edit},
		}

		provider := EditListProvider{}
		idx := buildLineOffsetIndex(content)

		resolved, err := provider.EditsWithLineIndex(content, idx, target)
		if err != nil {
			if resolved != nil {
				t.Fatalf("error result carries edits: %v alongside %d edits", err, len(resolved))
			}

			if !errors.Is(err, ErrEditStale) &&
				!errors.Is(err, ErrEditCrossFile) &&
				!errors.Is(err, ErrPositionUnresolvable) {
				t.Fatalf("unexpected error type: %v", err)
			}

			return
		}

		for _, fe := range resolved {
			if fe.Offset < 0 || fe.Offset > len(content) {
				t.Fatalf("resolved edit offset %d out of bounds for content len %d", fe.Offset, len(content))
			}
		}

		_ = applyEditsToContent(content, resolved)
	})
}
