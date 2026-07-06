package main

import (
	"bytes"
	"testing"

	. "github.com/onsi/gomega"
)

func TestToPipelineConfig_BadTimeout(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cfg := pipelineConfigFile{Timeout: "not-a-duration"}
	_, err := cfg.toPipelineConfig()
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("parse timeout"))
}

func TestToPipelineConfig_BadDetectorTimeout(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cfg := pipelineConfigFile{
		DetectorTimeouts: map[string]string{"govet": "bad"},
	}
	_, err := cfg.toPipelineConfig()
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("parse detector timeout"))
}

func TestToPipelineConfig_Success(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cfg := pipelineConfigFile{
		MaxIterations:    3,
		Timeout:          "30s",
		DetectorTimeouts: map[string]string{"govet": "5s"},
	}
	pc, err := cfg.toPipelineConfig()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(pc.MaxIterations).To(Equal(3))
	g.Expect(pc.DetectorTimeouts).To(HaveKey("govet"))
}

func TestOutputResults_Markdown(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	var buf bytes.Buffer
	report := reportWithFindings()

	err := outputResults(&buf, report, "markdown", true)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(buf.String()).To(ContainSubstring("nil dereference"))
}

func TestValidate_InvalidMaxIterations(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cfg := pipelineConfigFile{MaxIterations: -1}
	err := cfg.validate()
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("maxIterations"))
}
