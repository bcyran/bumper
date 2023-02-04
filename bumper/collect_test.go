package bumper

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/bcyran/bumper/internal/testutils"
	"github.com/bcyran/bumper/pack"
	"github.com/stretchr/testify/assert"
)

func createNamedPackage(path string, name string) {
	srcinfo := fmt.Sprintf(`
pkgbase = %s
        pkgname = %s
        url = some_url
        pkgver = some_ver
        pkgrel = some_rel
`, name, name)
	testutils.CreatePackage(path, []byte{}, []byte(srcinfo))
}

func packagesNames(packages []pack.Package) []string {
	names := []string{}
	for _, pack := range packages {
		names = append(names, pack.Pkgbase)
	}
	return names
}

func TestCollectPackages_Single(t *testing.T) {
	packagePath := filepath.Join(t.TempDir(), "pack")
	createNamedPackage(packagePath, "foo-pack")

	foundPackages, err := CollectPackages(packagePath, 0)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo-pack"}, packagesNames(foundPackages))
}

func TestCollectPackages_Recursive(t *testing.T) {
	rootDir := t.TempDir()
	createNamedPackage(filepath.Join(rootDir, ".a"), "ignore")
	createNamedPackage(filepath.Join(rootDir, "a"), "pack1")
	createNamedPackage(filepath.Join(rootDir, "b"), "pack2")
	createNamedPackage(filepath.Join(rootDir, "c/pack3"), "pack3")
	createNamedPackage(filepath.Join(rootDir, "d/more/pack4"), "pack4")
	ioutil.WriteFile(filepath.Join(rootDir, "random-file"), []byte("whatever"), 0o644)

	cases := map[int][]string{
		0: {},
		1: {"pack1", "pack2"},
		2: {"pack1", "pack2", "pack3"},
		3: {"pack1", "pack2", "pack3", "pack4"},
	}

	for depth, expectedNames := range cases {
		foundPackages, err := CollectPackages(rootDir, depth)
		assert.NoError(t, err)
		assert.Equal(t, expectedNames, packagesNames(foundPackages))
	}
}

func TestCollectPackages_Error(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "not-a-dir")
	ioutil.WriteFile(filePath, []byte{}, 0o644)

	cases := []int{0, 1}

	for _, depth := range cases {
		_, err := CollectPackages(filePath, depth)
		assert.ErrorIs(t, err, pack.ErrInvalidPath)
		assert.ErrorContains(t, err, "not a directory")
	}
}
