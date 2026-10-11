package finding

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"
)

var _ Detector = DetectorFunc(nil)

func TestDetectorFunc_Detect(t *testing.T) {
	g := NewParallelGomega(t)

	ctx := context.Background()
	want := []Finding{{Message: "found"}}

	fn := DetectorFunc(func(c context.Context) ([]Finding, error) {
		g.Expect(c).To(Equal(ctx))

		return want, nil
	})

	findings, err := fn.Detect(ctx)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(findings).To(Equal(want))
	g.Expect(fn.Name()).To(Equal(""))
}

func TestDetectorFunc_DetectError(t *testing.T) {
	g := NewParallelGomega(t)

	sentinel := errors.New("detect failed")
	fn := DetectorFunc(func(context.Context) ([]Finding, error) { return nil, sentinel })

	findings, err := fn.Detect(context.Background())
	g.Expect(err).To(MatchError(sentinel))
	g.Expect(findings).To(BeNil())
}

func TestNamedDetectorFunc(t *testing.T) {
	g := NewParallelGomega(t)

	want := []Finding{{Message: "found"}}
	d := NamedDetectorFunc("govet", func(context.Context) ([]Finding, error) { return want, nil })

	g.Expect(d.Name()).To(Equal("govet"))

	findings, err := d.Detect(context.Background())
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(findings).To(Equal(want))
}
