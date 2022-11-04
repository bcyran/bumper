package bumper

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"path/filepath"
)

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

func TestLoadPackage_MissingSrcinfoField(t *testing.T) {
	packagePath := filepath.Join(t.TempDir(), "package")
	srcinfo := []byte(`
pkgbase = foo
        pkgname = bar
        url = boo
        pkgrel =  baz
`)
	createPackage(packagePath, []byte{}, srcinfo)

	_, err := LoadPackage(packagePath)

	assert.ErrorIs(t, err, ErrInvalidSrcinfo)
	assert.ErrorContains(t, err, "missing 'pkgver' field")
}

func TestLoadPackage_CorrectSrcinfo(t *testing.T) {
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
		Path:    packagePath,
		Pkgname: "expected_name",
		Url:     "expected_url",
		Pkgver:  "expected_ver",
		Pkgrel:  "expected_rel",
	}
	assert.Equal(t, expectedPackage, *loadedPackage)
}

func createPackage(path string, pkgbuild []byte, srcinfo []byte) {
	os.MkdirAll(path, 0755)
	ioutil.WriteFile(filepath.Join(path, "PKGBUILD"), pkgbuild, 0644)
	ioutil.WriteFile(filepath.Join(path, ".SRCINFO"), srcinfo, 0644)
}
