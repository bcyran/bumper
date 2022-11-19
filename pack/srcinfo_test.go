package pack

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSrcinfo_Valid(t *testing.T) {
	srcinfoPath := filepath.Join(t.TempDir(), ".SRCINFO")
	srcinfoText := []byte(`
pkgbase = expected_base
        pkgname = expected_name
        url = expected_url
        pkgver = expected_ver
        pkgrel = expected_rel
	`)
	ioutil.WriteFile(srcinfoPath, srcinfoText, 0644)

	parsedSrcinfo, err := ParseSrcinfo(srcinfoPath)

	assert.NoError(t, err)
	expectedSrcinfo := Srcinfo{
		Pkgbase: "expected_base",
		Url:     "expected_url",
		FullVersion: &FullVersion{
			Pkgver: Version("expected_ver"),
			Pkgrel: "expected_rel",
		},
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
	ioutil.WriteFile(srcinfoPath, srcinfoText, 0644)

	_, err := ParseSrcinfo(srcinfoPath)

	assert.ErrorIs(t, err, ErrInvalidSrcinfo)
	assert.ErrorContains(t, err, "missing/invalid 'pkgver' value")
}
