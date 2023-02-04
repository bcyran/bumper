package pack

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bcyran/bumper/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	srcinfoBytes = []byte(`
pkgbase = expected_name
        pkgname = expected_name
        url = expected_url
        pkgver = expected_ver
        pkgrel = expected_rel
`)

	expectedSrcinfo = Srcinfo{
		Pkgbase: "expected_name",
		URL:     "expected_url",
		FullVersion: &FullVersion{
			Pkgver: "expected_ver",
			Pkgrel: "expected_rel",
		},
	}
)

func TestLoadPackage_Valid(t *testing.T) {
	packagePath := filepath.Join(t.TempDir(), "package")
	err := testutils.CreatePackage(packagePath, []byte{}, srcinfoBytes)
	require.Nil(t, err)

	loadedPackage, err := LoadPackage(packagePath)

	assert.NoError(t, err)
	expectedPackage := Package{
		Path:    packagePath,
		Srcinfo: &expectedSrcinfo,
		IsVCS:   false,
	}
	assert.Equal(t, expectedPackage, *loadedPackage)
}

func TestLoadPackage_ValidVCS(t *testing.T) {
	packagePath := filepath.Join(t.TempDir(), "package")
	pkgbuildBytes := []byte(`
pkgname = expected_name
pkgver() {
}
build() {
}
`)
	err := testutils.CreatePackage(packagePath, pkgbuildBytes, srcinfoBytes)
	require.Nil(t, err)

	loadedPackage, err := LoadPackage(packagePath)

	assert.NoError(t, err)
	expectedPackage := Package{
		Path:    packagePath,
		Srcinfo: &expectedSrcinfo,
		IsVCS:   true,
	}
	assert.Equal(t, expectedPackage, *loadedPackage)
}

func TestLoadPackage_PathNotExisting(t *testing.T) {
	notExistingPath := filepath.Join(t.TempDir(), "not_existing")

	_, err := LoadPackage(notExistingPath)

	assert.ErrorIs(t, err, ErrInvalidPath)
	assert.ErrorContains(t, err, "doesn't exist")
}

func TestLoadPackage_PathNotDirectory(t *testing.T) {
	notDirectoryPath := filepath.Join(t.TempDir(), "not_directory")
	_, createErr := os.Create(notDirectoryPath)
	require.Nil(t, createErr)

	_, err := LoadPackage(notDirectoryPath)

	assert.ErrorIs(t, err, ErrInvalidPath)
	assert.ErrorContains(t, err, "not a directory")
}

func TestLoadPackage_PathNoPkbuild(t *testing.T) {
	noPkgbuildPath := filepath.Join(t.TempDir(), "directory")
	mkdirErr := os.Mkdir(noPkgbuildPath, 0o755)
	require.Nil(t, mkdirErr)

	_, err := LoadPackage(noPkgbuildPath)

	assert.ErrorIs(t, err, ErrNotAPackage)
	assert.ErrorContains(t, err, "missing PKGBUILD")
}

func TestLoadPackage_PathNoSrcinfo(t *testing.T) {
	noSrcinfoPath := filepath.Join(t.TempDir(), "directory")
	mkdirErr := os.Mkdir(noSrcinfoPath, 0o755)
	require.Nil(t, mkdirErr)
	_, createErr := os.Create(filepath.Join(noSrcinfoPath, "PKGBUILD"))
	require.Nil(t, createErr)

	_, err := LoadPackage(noSrcinfoPath)

	assert.ErrorIs(t, err, ErrNotAPackage)
	assert.ErrorContains(t, err, "missing .SRCINFO")
}
