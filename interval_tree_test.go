package finding

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestIntervalTree_Empty(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	tree := NewIntervalIndex[int](nil)
	g.Expect(tree.Query(0, 10)).To(gomega.BeNil())
}

func TestIntervalTree_SingleInterval(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	tree := NewIntervalIndex([]Interval[string]{
		{Start: 5, End: 10, Value: "a"},
	})

	g.Expect(tree.Query(0, 5)).To(gomega.BeNil())     // before
	g.Expect(tree.Query(10, 15)).To(gomega.BeNil())   // after
	g.Expect(tree.Query(5, 10)).To(gomega.HaveLen(1)) // exact
	g.Expect(tree.Query(7, 8)).To(gomega.HaveLen(1))  // inside
	g.Expect(tree.Query(0, 6)).To(gomega.HaveLen(1))  // overlaps start
	g.Expect(tree.Query(9, 15)).To(gomega.HaveLen(1)) // overlaps end
}

func TestIntervalTree_MultipleOverlapping(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	tree := NewIntervalIndex([]Interval[string]{
		{Start: 0, End: 5, Value: "a"},
		{Start: 3, End: 8, Value: "b"},
		{Start: 6, End: 10, Value: "c"},
		{Start: 15, End: 20, Value: "d"},
	})

	g.Expect(tree.Query(0, 10)).To(gomega.HaveLen(3))  // a, b, c
	g.Expect(tree.Query(4, 7)).To(gomega.HaveLen(3))   // a=[0,5)✓, b=[3,8)✓, c=[6,10)✓
	g.Expect(tree.Query(14, 16)).To(gomega.HaveLen(1)) // d
	g.Expect(tree.Query(10, 15)).To(gomega.BeNil())    // gap
}

func TestIntervalTree_PointQuery(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	tree := NewIntervalIndex([]Interval[int]{
		{Start: 10, End: 20, Value: 1},
		{Start: 30, End: 40, Value: 2},
	})

	// Point query: [15, 16)
	g.Expect(tree.Query(15, 16)).To(gomega.HaveLen(1))
	g.Expect(tree.Query(25, 26)).To(gomega.BeNil())
}

func TestIntervalTree_AdjacentNoOverlap(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	tree := NewIntervalIndex([]Interval[string]{
		{Start: 0, End: 5, Value: "a"},
		{Start: 5, End: 10, Value: "b"},
	})

	// Adjacent: [0,5) and [5,10) do NOT overlap each other
	g.Expect(tree.Query(0, 5)).To(gomega.HaveLen(1))  // only a
	g.Expect(tree.Query(5, 10)).To(gomega.HaveLen(1)) // only b
	g.Expect(tree.Query(0, 10)).To(gomega.HaveLen(2)) // both
}
