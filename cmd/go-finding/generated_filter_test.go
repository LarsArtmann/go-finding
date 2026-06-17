package main

import (
	"testing"
	"time"

	"github.com/larsartmann/go-finding/pipeline"
	. "github.com/onsi/gomega"
)

func TestSplitCommaList(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", nil},
		{"single", "foo", []string{"foo"}},
		{"multiple", "foo,bar,baz", []string{"foo", "bar", "baz"}},
		{"with spaces", " foo , bar , baz ", []string{"foo", "bar", "baz"}},
		{"trailing comma", "foo,", []string{"foo"}},
		{"only commas", ",,,", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := splitCommaList(tt.input)
			g := NewWithT(t)
			g.Expect(got).To(Equal(tt.want))
		})
	}
}

func TestMustKeys(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	keys := mustKeys(filterTypeRegistry)
	g.Expect(keys).NotTo(BeEmpty())
	g.Expect(keys).To(ContainElement("all"))
	g.Expect(keys).To(ContainElement("sqlc"))
	g.Expect(keys).To(ContainElement("mockery"))
	g.Expect(keys).To(ContainElement("go-swagger"))
	g.Expect(keys).To(HaveLen(19))
}

func TestParseFilterGenTypes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		cliTypes   string
		configType string
		wantErr    bool
		wantLen    int
	}{
		{"all default", "", "", false, 1},
		{"cli all", "all", "", false, 1},
		{"single type", "sqlc", "", false, 1},
		{"multiple types", "sqlc,mockgen", "", false, 2},
		{"v3.2.0 new types", "counterfeiter,easyjson,ent,go-swagger,gqlgen,mockery,msgp", "", false, 7},
		{"config fallback", "", "protobuf", false, 1},
		{"unknown type", "nonexistent", "", true, 0},
		{"with spaces", " sqlc , mockgen ", "", false, 2},
		{"empty after trim", ",,", "", false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			opts, err := parseFilterGenTypes(tt.cliTypes, tt.configType)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}

			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(opts).To(HaveLen(tt.wantLen))
		})
	}
}

func TestAddGeneratedFilter_Integration(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	pipelineCfg := testPipelineConfig()
	cfg := pipelineConfigFile{}

	err := addGeneratedFilter(&pipelineCfg, cfg, "all", "", "")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(pipelineCfg.Processors).To(HaveLen(1))

	pipelineCfg2 := testPipelineConfig()

	err = addGeneratedFilter(&pipelineCfg2, cfg, "nonexistent-type", "", "")
	g.Expect(err).To(HaveOccurred())
}

func TestAddGeneratedFilter_WithExcludeInclude(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	pipelineCfg := testPipelineConfig()
	cfg := pipelineConfigFile{
		GeneratedExclude: []string{"vendor/*"},
		GeneratedInclude: []string{"src/**/*.go"},
	}

	err := addGeneratedFilter(&pipelineCfg, cfg, "all", "gen/*", "src/*")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(pipelineCfg.Processors).To(HaveLen(1))
}

func testPipelineConfig() pipeline.Config {
	return pipeline.Config{ //nolint:exhaustruct
		MaxIterations:     1,
		ParallelDetectors: false,
		Timeout:           30 * time.Second,
	}
}
