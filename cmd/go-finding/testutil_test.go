package main

import (
	"testing"

	"github.com/onsi/gomega"
)

// NewParallelGomega marks the test as parallel and returns a gomega.GomegaWithT,
// consolidating the t.Parallel() + gomega.NewWithT(t) boilerplate.
func NewParallelGomega(t *testing.T) *gomega.GomegaWithT {
	t.Helper()
	t.Parallel()

	return gomega.NewWithT(t)
}
