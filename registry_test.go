package finding

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
)

func stubDetector(name string) *namedDetector {
	return &namedDetector{name: name, fn: func(context.Context) ([]Finding, error) { return nil, nil }}
}

func TestDetectorRegistry_Build(t *testing.T) {
	g := NewParallelGomega(t)

	r := NewDetectorRegistry()
	g.Expect(r.Register("govet", func() Detector { return stubDetector("govet") })).To(Succeed())
	g.Expect(r.Register("govet", func() Detector { return stubDetector("govet") })).
		To(MatchError(ErrDetectorRegistered))

	d, err := r.Build("govet")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(d.Name()).To(Equal("govet"))

	_, err = r.Build("nope")
	g.Expect(err).To(MatchError(ErrUnknownDetector))
}

func TestDetectorRegistry_MustRegister_PanicsOnDuplicate(t *testing.T) {
	g := NewParallelGomega(t)

	r := NewDetectorRegistry()
	r.MustRegister("govet", func() Detector { return stubDetector("govet") })

	g.Expect(func() {
		r.MustRegister("govet", func() Detector { return stubDetector("govet") })
	}).To(Panic())
}

func TestDetectorRegistry_BuildAll_Sorted(t *testing.T) {
	g := NewParallelGomega(t)

	r := NewDetectorRegistry()
	r.MustRegister("zeta", func() Detector { return stubDetector("zeta") })
	r.MustRegister("alpha", func() Detector { return stubDetector("alpha") })
	r.MustRegister("mid", func() Detector { return stubDetector("mid") })

	ds, err := r.BuildAll()
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ds).To(HaveLen(3))

	names := []string{ds[0].Name(), ds[1].Name(), ds[2].Name()}
	g.Expect(names).To(Equal([]string{"alpha", "mid", "zeta"}))
}

func TestDetectorRegistry_BuildAll_Empty(t *testing.T) {
	g := NewParallelGomega(t)

	ds, err := NewDetectorRegistry().BuildAll()
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ds).To(BeEmpty())
}
