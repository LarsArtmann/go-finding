package finding

import (
	"bytes"
	"io"
	"testing"
)

type probeCounter struct {
	w io.Writer
	n int
}

func (c *probeCounter) Write(p []byte) (int, error) { c.n++; return c.w.Write(p) }

func TestProbeWrites(t *testing.T) {
	var buf bytes.Buffer
	cw := &probeCounter{w: &buf}
	err := nanConfidenceFinding().WriteJSON(cw)
	t.Logf("err=%v writes=%d buflen=%d", err, cw.n, buf.Len())
}
