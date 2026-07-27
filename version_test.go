package coverage_test

import (
	"testing"

	coverage "coderaiser/go-coverage"

	tape "github.com/coderaiser/go-tape"
)

func TestVersion(t *testing.T) {
	tape.Test(t, "version: VersionFromJSON returns version string", func(t *tape.T) {
		result := coverage.VersionFromJSON([]byte(`{"version":"1.2.3"}`))
		t.Equal(result, "1.2.3")
		t.End()
	})

	tape.Test(t, "version: VersionFromJSON returns unknown on invalid JSON", func(t *tape.T) {
		result := coverage.VersionFromJSON([]byte(`{invalid`))
		t.Equal(result, "unknown")
		t.End()
	})

	tape.Test(t, "version: VersionFromJSON returns unknown on empty version", func(t *tape.T) {
		result := coverage.VersionFromJSON([]byte(`{"version":""}`))
		t.Equal(result, "unknown")
		t.End()
	})

	tape.Test(t, "version: VersionLine contains binary name", func(t *tape.T) {
		result := coverage.VersionLine()
		t.Equal(result, "v1.1.0")
		t.End()
	})
}
