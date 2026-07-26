package coverage

import (
	"encoding/json"
	"fmt"
)

// Version returns the version string embedded from package.json at build time.
func VersionFromJSON(packageJSONBytes []byte) string {
	var pkg struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(packageJSONBytes, &pkg); err != nil {
		return "unknown"
	}
	if pkg.Version == "" {
		return "unknown"
	}
	return pkg.Version
}

func Version() string {
	return VersionFromJSON(packageJSONBytes)
}

// VersionLine returns "go-coverage x.y.z" for -v output.
func VersionLine() string {
	return fmt.Sprintf("coverage %s", Version())
}
