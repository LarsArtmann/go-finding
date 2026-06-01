package detectors

import "path/filepath"

// DetectorName constants are the single source of truth for detector names.
const (
	DetectorNameGovet       = "govet"
	DetectorNameStaticcheck = "staticcheck"
)

// resolvePath joins dir with file if file is not already absolute.
func resolvePath(dir, file string) string {
	if dir != "" && !filepath.IsAbs(file) {
		return filepath.Join(dir, file)
	}

	return file
}
