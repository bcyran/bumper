package pack

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSrcinfo_Valid(t *testing.T) {
	srcinfoPath := filepath.Join(t.TempDir(), ".SRCINFO")
	srcinfoText := []byte(`
pkgbase = expected_base
        pkgname = expected_name
        url = expected_url
        pkgver     = expected_ver
        pkgrel =     expected_rel
		source = https://fake.source
		source = baz::https://foo.bar
	`)
	writeErr := os.WriteFile(srcinfoPath, srcinfoText, 0o644)
	require.Nil(t, writeErr)

	parsedSrcinfo, err := ParseSrcinfo(srcinfoPath)

	assert.NoError(t, err)
	expectedSrcinfo := Srcinfo{
		Pkgbase: "expected_base",
		URL:     "expected_url",
		FullVersion: &FullVersion{
			Pkgver: Version("expected_ver"),
			Pkgrel: "expected_rel",
		},
		Source: []string{"https://fake.source", "baz::https://foo.bar"},
	}
	assert.Equal(t, expectedSrcinfo, *parsedSrcinfo)
}

func TestParseSrcinfo_Invalid(t *testing.T) {
	srcinfoPath := filepath.Join(t.TempDir(), ".SRCINFO")
	srcinfoText := []byte(`
pkgbase = expected_base
        pkgname = expected_name
        url = expected_url
        pkgrel = expected_rel
	`)
	writeErr := os.WriteFile(srcinfoPath, srcinfoText, 0o644)
	require.Nil(t, writeErr)

	_, err := ParseSrcinfo(srcinfoPath)

	assert.ErrorIs(t, err, ErrInvalidSrcinfo)
	assert.ErrorContains(t, err, "missing/invalid 'pkgver' value")
}
