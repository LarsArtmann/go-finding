package finding

import (
	"context"
	"errors"
	"testing"
)

func TestTextEdit_HasSpan(t *testing.T) {
	g := NewParallelGomega(t)

	replacement := TextEdit{
		Start:   Pos("a.go", 1, 1),
		End:     Pos("a.go", 1, 5),
		NewText: "x",
	}
	g.Expect(replacement.HasSpan()).To(BeTrue())
	g.Expect(replacement.IsInsertion()).To(BeFalse())
	g.Expect(replacement.IsDeletion()).To(BeFalse())

	insertion := TextEdit{Start: Pos("a.go", 1, 1), NewText: "x"}
	g.Expect(insertion.HasSpan()).To(BeFalse())
	g.Expect(insertion.IsInsertion()).To(BeTrue())

	deletion := TextEdit{Start: Pos("a.go", 1, 1), End: Pos("a.go", 1, 5)}
	g.Expect(deletion.IsDeletion()).To(BeTrue())
}

func TestTextEdit_EffectiveFile(t *testing.T) {
	g := NewParallelGomega(t)

	g.Expect(TextEdit{Start: Pos("b.go", 1, 1)}.EffectiveFile("a.go")).To(Equal(FilePath("b.go")))
	g.Expect(TextEdit{Start: Position{Line: 1, Offset: -1}}.EffectiveFile("a.go")).To(Equal(FilePath("a.go")))
}

func TestTextEdit_Validate(t *testing.T) {
	tests := []struct {
		name    string
		edit    TextEdit
		wantErr bool
	}{
		{
			name: "replacement with offsets is valid",
			edit: TextEdit{
				Start:   Pos("a.go", 1, 1),
				End:     Pos("a.go", 1, 5),
				NewText: "x",
			},
		},
		{
			name:    "insertion without span is valid",
			edit:    TextEdit{Start: Pos("a.go", 1, 1), NewText: "x"},
			wantErr: false,
		},
		{
			name: "deletion without replacement is valid",
			edit: TextEdit{
				Start: Pos("a.go", 1, 1),
				End:   Pos("a.go", 2, 1),
			},
		},
		{
			name: "zero start position means byte zero and is valid",
			edit: TextEdit{Start: Position{}, NewText: "x"},
		},
		{
			name:    "unset start position is invalid",
			edit:    TextEdit{Start: Position{Offset: -1}, NewText: "x"},
			wantErr: true,
		},
		{
			name: "end offset before start offset is invalid",
			edit: TextEdit{
				Start: Position{File: "a.go", Line: 2, Offset: 20},
				End:   Position{File: "a.go", Line: 2, Offset: 10},
			},
			wantErr: true,
		},
		{
			name: "end line before start line is invalid",
			edit: TextEdit{
				Start: Pos("a.go", 5, 1),
				End:   Pos("a.go", 3, 1),
			},
			wantErr: true,
		},
		{
			name: "end column before start column on same line is invalid",
			edit: TextEdit{
				Start: Pos("a.go", 5, 9),
				End:   Pos("a.go", 5, 2),
			},
			wantErr: true,
		},
		{
			name: "end in different file is invalid",
			edit: TextEdit{
				Start: Pos("a.go", 1, 1),
				End:   Pos("b.go", 1, 5),
			},
			wantErr: true,
		},
		{
			name: "start offset with line-only end is invalid",
			edit: TextEdit{
				Start: Position{File: "a.go", Line: 1, Offset: 10},
				End:   Pos("a.go", 1, 5),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewParallelGomega(t)
			if tt.wantErr {
				g.Expect(tt.edit.Validate()).To(HaveOccurred())
			} else {
				g.Expect(tt.edit.Validate()).NotTo(HaveOccurred())
			}
		})
	}
}

func TestTextEdit_EqualAndCompare(t *testing.T) {
	g := NewParallelGomega(t)

	a := TextEdit{Start: Pos("a.go", 1, 1), End: Pos("a.go", 1, 5), NewText: "x"}
	same := a
	other := TextEdit{Start: Pos("a.go", 1, 1), End: Pos("a.go", 1, 6), NewText: "x"}

	g.Expect(a.Equal(same)).To(BeTrue())
	g.Expect(a.Equal(other)).To(BeFalse())
	g.Expect(a.Compare(other)).NotTo(Equal(0))
}

func TestEditsEqual_OrderIndependent(t *testing.T) {
	g := NewParallelGomega(t)

	first := TextEdit{
		Start: Position{File: "a.go", Line: 2, Offset: 10},
		End:   Position{File: "a.go", Line: 2, Offset: 15},
	}
	second := TextEdit{
		Start: Position{File: "a.go", Line: 4, Offset: 30},
		End:   Position{File: "a.go", Line: 4, Offset: 32},
	}

	g.Expect(editsEqual([]TextEdit{first, second}, []TextEdit{first, second})).To(BeTrue())
	g.Expect(editsEqual([]TextEdit{first, second}, []TextEdit{second, first})).To(BeTrue())
	g.Expect(editsEqual([]TextEdit{first}, []TextEdit{first, second})).To(BeFalse())
	g.Expect(editsEqual(nil, nil)).To(BeTrue())
}

func TestFinding_Equal_WithEdits(t *testing.T) {
	g := NewParallelGomega(t)

	base := func(edits []TextEdit) Finding {
		return Finding{
			ID:          "id",
			Rule:        "r",
			ToolName:    "t",
			Message:     "m",
			Severity:    SeverityError,
			Position:    Pos("a.go", 1, 1),
			FixStrategy: FixStrategyDirect,
			Edits:       edits,
		}
	}

	first := TextEdit{Start: Pos("a.go", 2, 1), End: Pos("a.go", 2, 5)}
	second := TextEdit{Start: Pos("a.go", 4, 1), NewText: "inserted"}

	g.Expect(base([]TextEdit{first, second}).Equal(base([]TextEdit{first, second}))).To(BeTrue())
	g.Expect(base([]TextEdit{first, second}).Equal(base([]TextEdit{second, first}))).To(BeTrue())
	g.Expect(base([]TextEdit{first}).Equal(base([]TextEdit{first, second}))).To(BeFalse())
}

func TestBuilder_WithEdits(t *testing.T) {
	g := NewParallelGomega(t)

	edit := TextEdit{Start: Pos("a.go", 2, 1), End: Pos("a.go", 2, 5), NewText: "fixed"}

	f := NewBuilder("r", "t", "m", SeverityWarning, Pos("a.go", 1, 1)).
		WithFixStrategy(FixStrategyDirect).
		WithEdits(edit).
		BuildOrDefault()

	g.Expect(f.Edits).To(HaveLen(1))
	g.Expect(f.Edits[0].Equal(edit)).To(BeTrue())
	g.Expect(f.HasEditList()).To(BeTrue())
	g.Expect(f.HasCodeChange()).To(BeTrue())
	g.Expect(f.HasFix()).To(BeTrue())
	g.Expect(f.IsAutoFixable()).To(BeTrue())
}

func TestFinding_FixPredicates_WithEditsOnly(t *testing.T) {
	g := NewParallelGomega(t)

	f := Finding{
		ID:          "id",
		Rule:        "r",
		ToolName:    "t",
		Message:     "m",
		Severity:    SeverityError,
		Position:    Pos("a.go", 1, 1),
		FixStrategy: FixStrategyDirect,
		Edits:       []TextEdit{{Start: Pos("a.go", 1, 1), NewText: "x"}},
	}

	g.Expect(f.HasCodeChange()).To(BeTrue())
	g.Expect(f.HasFix()).To(BeTrue())
	g.Expect(f.IsAutoFixable()).To(BeTrue())
	g.Expect(f.Validate()).NotTo(HaveOccurred())
}

func TestFinding_JSONRoundTrip_WithEdits(t *testing.T) {
	g := NewParallelGomega(t)

	f := Finding{
		ID:          "id",
		Rule:        "r",
		ToolName:    "t",
		Message:     "m",
		Severity:    SeverityError,
		Position:    Pos("a.go", 1, 1),
		FixStrategy: FixStrategyDirect,
		Edits: []TextEdit{
			{Start: Pos("a.go", 2, 1), End: Pos("a.go", 2, 5), NewText: "fixed"},
			{Start: Pos("a.go", 4, 1), NewText: "inserted"},
		},
	}

	data, err := marshalJSONString(f)
	g.Expect(err).NotTo(HaveOccurred())

	var restored Finding
	g.Expect(unmarshalFinding(data, &restored)).To(Succeed())
	g.Expect(restored.Equal(f)).To(BeTrue(), "edits must survive JSON round-trip: %s", data)
}

func TestSARIF_RoundTrip_WithEdits(t *testing.T) {
	g := NewParallelGomega(t)

	f := Finding{
		ID:          "id",
		Rule:        "r",
		ToolName:    "t",
		Message:     "m",
		Severity:    SeverityError,
		Position:    Pos("a.go", 1, 1),
		FixStrategy: FixStrategyDirect,
		Suggestion:  "rewrite",
		BeforeCode:  "old",
		AfterCode:   "new",
		Edits: []TextEdit{
			{
				Start:   Position{File: "a.go", Line: 2, Column: 1, Offset: 10},
				End:     Position{File: "a.go", Line: 2, Column: 5, Offset: 14},
				NewText: "new",
			},
			{
				Start:   Position{File: "a.go", Line: 4, Column: 1, Offset: 0},
				End:     Position{File: "a.go", Line: 4, Column: 1, Offset: 0},
				NewText: "inserted",
			},
		},
	}

	report := NewReportFromFindings(ToolInfo{Name: "t"}, []Finding{f})

	data, err := report.ToSARIF()
	g.Expect(err).NotTo(HaveOccurred())

	imported, err := FindingsFromSARIF(context.Background(), data)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(imported).To(HaveLen(1))

	got := imported[0]
	g.Expect(got.Edits).To(HaveLen(2))
	g.Expect(got.Edits).To(ConsistOf(f.Edits[0], f.Edits[1]))
	g.Expect(got.AfterCode).To(Equal("new"))
	g.Expect(got.FixStrategy).To(Equal(FixStrategyDirect))
}

func TestTextEdit_Validate_SentinelErrors(t *testing.T) {
	g := NewParallelGomega(t)

	inverted := TextEdit{
		Start: Position{File: "a.go", Line: 2, Offset: 20},
		End:   Position{File: "a.go", Line: 2, Offset: 10},
	}

	err := inverted.Validate()
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, errEditInverted)).To(BeTrue())

	crossFile := TextEdit{
		Start: Pos("a.go", 1, 1),
		End:   Pos("b.go", 1, 5),
	}

	err = crossFile.Validate()
	g.Expect(err).To(HaveOccurred())
	g.Expect(errors.Is(err, errEditCrossFile)).To(BeTrue())
}
