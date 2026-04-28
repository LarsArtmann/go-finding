// basic demonstrates creating a Finding and Report from scratch.
package main

import (
	"fmt"
	"log"

	finding "github.com/larsartmann/go-finding"
)

func main() {
	f := finding.Finding{
		ID:       "govet:printf:main.go:42:10",
		Rule:     "printf",
		ToolName: "govet",
		Message:  "Printf format %s has arg 42 of wrong type int",
		Severity: finding.SeverityError,
		Position: finding.Position{File: "main.go", Line: 42, Column: 10},
		Category: finding.CategoryCorrectness,
	}

	r := finding.NewReport(finding.ToolInfo{Name: "govet", Version: "1.23"})
	r.AddFinding(f)
	r.ComputeSummary()

	fmt.Printf("Report has %d finding(s)\n", r.Summary.Total)
	fmt.Printf("Severity: %s\n", f.Severity)

	json, err := r.PrettyJSON()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(json)
}
