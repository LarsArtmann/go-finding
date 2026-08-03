module github.com/larsartmann/go-finding/cmd/go-finding

go 1.26.5

require (
	github.com/LarsArtmann/gogenfilter/v3 v3.3.2
	github.com/go-faster/yaml v0.4.6
	github.com/larsartmann/go-finding v0.0.0-00010101000000-000000000000
	github.com/larsartmann/go-finding/pipeline v0.0.0-00010101000000-000000000000
	github.com/larsartmann/go-output v0.36.0
	github.com/larsartmann/go-output/delimited v0.35.0
	github.com/larsartmann/go-output/markdown v0.35.0
	github.com/onsi/gomega v1.42.1
)

require (
	github.com/bmatcuk/doublestar/v4 v4.10.0 // indirect
	github.com/go-faster/errors v0.8.0 // indirect
	github.com/go-faster/jx v1.2.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/larsartmann/go-branded-id v0.5.1 // indirect
	github.com/larsartmann/go-error-family v0.10.0 // indirect
	github.com/larsartmann/go-output/escape v0.35.0 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/term v0.45.0 // indirect
	golang.org/x/text v0.40.0 // indirect
)

replace (
	github.com/larsartmann/go-finding => ../..
	github.com/larsartmann/go-finding/pipeline => ../../pipeline
)
