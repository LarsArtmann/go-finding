package lockutil_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/larsartmann/go-finding/lockutil"
)

func TestLocked_IncrementsCounter(t *testing.T) {
	t.Parallel()

	var (
		mu sync.Mutex
		n  int
	)

	lockutil.Locked(&mu, func() struct{} {
		n++

		return struct{}{}
	})

	if n != 1 {
		t.Fatalf("expected n=1, got %d", n)
	}
}

func TestLocked_ReturnsValue(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex

	got := lockutil.Locked(&mu, func() int { return 42 })

	if got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
}

func TestRLocked_ReturnsValue(t *testing.T) {
	t.Parallel()

	var (
		mu  sync.RWMutex
		val = 7
	)

	got := lockutil.RLocked(&mu, func() int { return val })

	if got != 7 {
		t.Fatalf("expected 7, got %d", got)
	}
}

func TestLocked_ConcurrentSafe(t *testing.T) {
	t.Parallel()

	var (
		mu sync.Mutex
		n  atomic.Int64
	)

	const goroutines = 100

	var wg sync.WaitGroup

	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()

			lockutil.Locked(&mu, func() struct{} {
				n.Add(1)

				return struct{}{}
			})
		}()
	}

	wg.Wait()

	if got := n.Load(); got != goroutines {
		t.Fatalf("expected %d increments, got %d", goroutines, got)
	}
}
