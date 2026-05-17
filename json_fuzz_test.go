package finding

import (
	"testing"
)

func FuzzFindingsFromJSON(f *testing.F) {
	f.Add(
		[]byte(
			`[{"id":"f1","rule":"r1","toolName":"t","message":"m","severity":"error","position":{"file":"a.go","line":1}}]`,
		),
	)
	f.Add([]byte(`[{"id":"","rule":"","message":""}]`))
	f.Add([]byte(`not json`))
	f.Add([]byte{})
	f.Add([]byte(`{"id":"f1"}`))
	f.Add([]byte(`[{"id":"f1","rule":"r1"`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`null`))
	f.Add([]byte(`[null]`))

	f.Fuzz(func(_ *testing.T, data []byte) {
		_, _, _ = FindingsFromJSON(data)
	})
}

func FuzzReportFromJSON(f *testing.F) {
	f.Add(
		[]byte(
			`{"tool":{"name":"t"},"findings":[{"id":"f1","rule":"r1","toolName":"t","message":"m","severity":"error","position":{"file":"a.go","line":1}}]}`,
		),
	)
	f.Add([]byte(`{"tool":{"name":""}}`))
	f.Add([]byte(`not json`))
	f.Add([]byte{})
	f.Add([]byte(`{"tool":{"name":"t"},"findings":[{"id":"","rule":"","message":""}]}`))
	f.Add([]byte(`null`))

	f.Fuzz(func(_ *testing.T, data []byte) {
		_, _, _ = ReportFromJSON(data)
	})
}

func FuzzFromJSON(f *testing.F) {
	f.Add(
		[]byte(
			`{"id":"f1","rule":"r1","toolName":"t","message":"m","severity":"error","position":{"file":"a.go","line":1}}`,
		),
	)
	f.Add([]byte(`{"id":"","rule":"","message":""}`))
	f.Add([]byte(`not json`))
	f.Add([]byte{})
	f.Add([]byte(`null`))
	f.Add([]byte(`{"id":"f1"`))

	f.Fuzz(func(_ *testing.T, data []byte) {
		_, _ = FromJSON(data)
	})
}
