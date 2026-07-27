package coverage_test

import (
	"testing"

	coverage "coderaiser/go-coverage"
	tape "github.com/coderaiser/go-tape"
)

func TestVersion(t *testing.T) {
	tape.Test(t, "version: VersionFromJSON returns version string", func(t *tape.T) {
		t.Equal(coverage.VersionFromJSON([]byte(`{"version":"1.2.3"}`)), "1.2.3")
		t.End()
	})

	tape.Test(t, "version: VersionFromJSON returns unknown on invalid JSON", func(t *tape.T) {
		t.Equal(coverage.VersionFromJSON([]byte(`{invalid`)), "unknown")
		t.End()
	})

	tape.Test(t, "version: VersionFromJSON returns unknown on empty version", func(t *tape.T) {
		t.Equal(coverage.VersionFromJSON([]byte(`{"version":""}`)), "unknown")
		t.End()
	})

	tape.Test(t, "version: VersionLine contains binary name", func(t *tape.T) {
		t.Match(coverage.VersionLine(), "coverage ")
		t.End()
	})
}
