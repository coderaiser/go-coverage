package coverage_test

import (
	"testing"

	"coderaiser/go-coverage"
	"coderaiser/go-coverage/internal/assert"
)

func TestVersionFromJSONReturnsVersion(t *testing.T) {
	got := coverage.VersionFromJSON([]byte(`{"version":"1.2.3"}`))

	assert.Equal(t, "1.2.3", got)
}

func TestVersionFromJSONReturnsUnknownOnInvalidJSON(t *testing.T) {
	got := coverage.VersionFromJSON([]byte(`{invalid`))

	assert.Equal(t, "unknown", got)
}

func TestVersionFromJSONReturnsUnknownOnEmptyVersion(t *testing.T) {
	got := coverage.VersionFromJSON([]byte(`{"version":""}`))

	assert.Equal(t, "unknown", got)
}

func TestVersionLine(t *testing.T) {
	got := coverage.VersionLine()

	assert.Contains(t, got, "coverage ")
}
