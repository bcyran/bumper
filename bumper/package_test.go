package bumper

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"path/filepath"
)

func TestLoadPackage_Valid(t *testing.T) {
	packagePath := filepath.Join(t.TempDir(), "package")
	srcinfo := []byte(`
pkgbase = expected_name
        pkgname = expected_name
        url = expected_url
        pkgver = expected_ver
        pkgrel = expected_rel
`)
	createPackage(packagePath, []byte{}, srcinfo)

	loadedPackage, err := LoadPackage(packagePath)

	assert.NoError(t, err)
	expectedPackage := Package{
		Path: packagePath,
		Srcinfo: &Srcinfo{
			Pkgbase: "expected_name",
			Url:     "expected_url",
			FullVersion: &FullVersion{
				Pkgver: "expected_ver",
				Pkgrel: "expected_rel",
			},
		},
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
	os.Mkdir(noPkgbuildPath, 0755)

	_, err := LoadPackage(noPkgbuildPath)

	assert.ErrorIs(t, err, ErrNotAPackage)
	assert.ErrorContains(t, err, "missing PKGBUILD")
}

func TestLoadPackage_PathNoSrcinfo(t *testing.T) {
	noSrcinfoPath := filepath.Join(t.TempDir(), "directory")
	os.Mkdir(noSrcinfoPath, 0755)
	os.Create(filepath.Join(noSrcinfoPath, "PKGBUILD"))

	_, err := LoadPackage(noSrcinfoPath)

	assert.ErrorIs(t, err, ErrNotAPackage)
	assert.ErrorContains(t, err, "missing .SRCINFO")
}

func createPackage(path string, pkgbuild []byte, srcinfo []byte) {
	os.MkdirAll(path, 0755)
	ioutil.WriteFile(filepath.Join(path, "PKGBUILD"), pkgbuild, 0644)
	ioutil.WriteFile(filepath.Join(path, ".SRCINFO"), srcinfo, 0644)
}
