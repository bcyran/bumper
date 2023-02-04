package pack

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bcyran/bumper/internal/testutils"
	"github.com/stretchr/testify/assert"
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
	testutils.CreatePackage(packagePath, []byte{}, srcinfoBytes)

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
	testutils.CreatePackage(packagePath, pkgbuildBytes, srcinfoBytes)

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
	os.Create(notDirectoryPath)

	_, err := LoadPackage(notDirectoryPath)

	assert.ErrorIs(t, err, ErrInvalidPath)
	assert.ErrorContains(t, err, "not a directory")
}

func TestLoadPackage_PathNoPkbuild(t *testing.T) {
	noPkgbuildPath := filepath.Join(t.TempDir(), "directory")
	os.Mkdir(noPkgbuildPath, 0o755)

	_, err := LoadPackage(noPkgbuildPath)

	assert.ErrorIs(t, err, ErrNotAPackage)
	assert.ErrorContains(t, err, "missing PKGBUILD")
}

func TestLoadPackage_PathNoSrcinfo(t *testing.T) {
	noSrcinfoPath := filepath.Join(t.TempDir(), "directory")
	os.Mkdir(noSrcinfoPath, 0o755)
	os.Create(filepath.Join(noSrcinfoPath, "PKGBUILD"))

	_, err := LoadPackage(noSrcinfoPath)

	assert.ErrorIs(t, err, ErrNotAPackage)
	assert.ErrorContains(t, err, "missing .SRCINFO")
}
