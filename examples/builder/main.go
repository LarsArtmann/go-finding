// builder demonstrates the fluent Finding builder API.
package main

import (
	"fmt"
	"log"

	finding "github.com/larsartmann/go-finding"
)

func main() {
	f := finding.NewBuilder("staticcheck", "SA1000", "invalid regular expression", finding.SeverityError, finding.Pos("pkg/validate.go", 24, 8)).
		WithCategory(finding.CategoryCorrectness).
		WithConfidence(0.95).
		WithBeforeCode("oldPattern").
		WithAfterCode("newPattern").
		WithFixStrategy(finding.FixStrategyDirect).
		Build()

	fmt.Printf("Finding: %s (%s)\n", f.ID, f.Rule)
	fmt.Printf("Position: %s\n", f.Position)
	fmt.Printf("Confidence: %.2f\n", f.Confidence)

	if f.HasFix() {
		fmt.Printf("Fix available: %s → %s\n", f.BeforeCode, f.AfterCode)
	}

	json, err := f.LineJSON()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(json)
}
