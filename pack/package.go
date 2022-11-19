package pack

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bcyran/bumper/upstream"
)

var (
	ErrInvalidPath = errors.New("invalid package path")
	ErrNotAPackage = errors.New("not a package")
)

type Package struct {
	*Srcinfo
	Path            string
	UpstreamVersion upstream.Version
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
func pkgbuildPath(path string) string {
	return filepath.Join(path, "PKGBUILD")
}

// srcinfoPath returns path to .SRCINFO given package root path.
func srcinfoPath(path string) string {
	return filepath.Join(path, ".SRCINFO")
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

	return &Package{Path: absPath, Srcinfo: srcinfo}, nil
}
