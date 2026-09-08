// Command fakestcheck is a test fixture masquerading as `staticcheck` for
// CLI e2e tests. It emits one staticcheck-format JSON line carrying the
// go-finding before/after fix extension, so findings parsed by the real
// CLI's staticcheck detector are auto-fixable and drive the fix pipeline
// end to end. Built by e2e_test.go and placed first on PATH.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println(`{"code":"ST1000","severity":"warning","location":{"file":"main.go","line":4,"column":2},"message":"old() should be new()","before":"old()","after":"new()"}`)
	os.Exit(1) // staticcheck convention: nonzero when findings exist
}
