package pack

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bcyran/bumper/upstream"
)

var (
	ErrInvalidPath = errors.New("invalid package path")
	ErrNotAPackage = errors.New("not a package")

	vcsToken = "pkgver()"
)

type Package struct {
	*Srcinfo
	Path            string
	UpstreamVersion upstream.Version
	IsOutdated      bool
	IsVCS           bool
}

func (pkg *Package) PkgbuildPath() string {
	return pkgbuildPath(pkg.Path)
}

func (pkg *Package) SrcinfoPath() string {
	return srcinfoPath(pkg.Path)
}

// LoadPackage tries to create Package struct based on given package dir path.
func LoadPackage(path string) (*Package, error) {
	if err := validateIsDir(path); err != nil {
		return &Package{}, err
	}
	if err := validateIsPackage(path); err != nil {
		return &Package{}, err
	}
	return makePackage(path)
}

// validateIsDir checks whatever given path leads to an existing directory.
func validateIsDir(path string) error {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("%w: doesn't exist or not accessible", ErrInvalidPath)
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("%w: not a directory", ErrInvalidPath)
	}
	return nil
}

// validateIsPackage checks whatever the directory at given path is an AUR package.
func validateIsPackage(path string) error {
	if _, err := os.Stat(pkgbuildPath(path)); os.IsNotExist(err) {
		return fmt.Errorf("%w: missing PKGBUILD", ErrNotAPackage)
	}
	if _, err := os.Stat(srcinfoPath(path)); os.IsNotExist(err) {
		return fmt.Errorf("%w: missing .SRCINFO", ErrNotAPackage)
	}
	return nil
}

// pkgbuildPath returns path to PKGBUILD given package root path.
func pkgbuildPath(pkgPath string) string {
	return filepath.Join(pkgPath, "PKGBUILD")
}

// srcinfoPath returns path to .SRCINFO given package root path.
func srcinfoPath(pkgPath string) string {
	return filepath.Join(pkgPath, ".SRCINFO")
}

// makePackage creates Package struct based on given package path dir without any safety checks.
func makePackage(path string) (*Package, error) {
	srcinfo, err := ParseSrcinfo(srcinfoPath(path))
	if err != nil {
		return &Package{}, err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return &Package{}, err
	}

	isVCS, err := isVCS(pkgbuildPath(path))
	if err != nil {
		return &Package{}, err
	}

	return &Package{Path: absPath, Srcinfo: srcinfo, IsVCS: isVCS}, nil
}

// isVCS checks if package is a VCS package (PKGBUILD contains 'pkgver()').
func isVCS(pkgbuildPath string) (bool, error) {
	pkgBuildBytes, err := ioutil.ReadFile(pkgbuildPath)
	if err != nil {
		return false, err
	}
	pkgBuild := string(pkgBuildBytes)

	return strings.Contains(pkgBuild, vcsToken), nil
}
