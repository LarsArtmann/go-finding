package toolsdk

// Trigger constructors for common tool shapes.

// OnGoFiles returns a Trigger matching Go source files in the Go language.
// The canonical trigger for Go linters, formatters, and analyzers.
func OnGoFiles() Trigger {
	return Trigger{
		Files:    []string{"**/*.go"},
		Language: "go",
	}
}

// OnGoModule returns a Trigger matching Go source files that requires a Go
// module file (go.mod or go.work). Use for tools that need module context
// to function (e.g. go-fix, go-structure-linter, govulncheck).
func OnGoModule() Trigger {
	return Trigger{
		Files:    []string{"**/*.go"},
		Language: "go",
		Requires: []string{"**/go.mod", "**/go.work"},
	}
}

// OnFiles returns a Trigger matching the given language and file patterns.
// Pass "" for language to match any language (language-agnostic tools).
func OnFiles(language string, patterns ...string) Trigger {
	return Trigger{
		Files:    patterns,
		Language: language,
	}
}

// AnyLanguage returns a Trigger that is language-agnostic but activates on the
// given file patterns. Use for tools that span languages (e.g. a generic
// task-marker scanner).
func AnyLanguage(patterns ...string) Trigger {
	return Trigger{
		Files:    patterns,
		Language: "",
	}
}
