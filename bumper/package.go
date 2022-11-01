package bumper

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const srcinfoSeparator = " = "

type Package struct {
	Path    string
	Pkgname string
	Url     string
	Pkgver  string
	Pkgrel  string
}

// Try to create `Package` struct based on given package dir path
func LoadPackage(path string) (*Package, error) {
	if err := validateIsDir(path); err != nil {
		return &Package{}, err
	}
	if err := validateIsPackage(path); err != nil {
		return &Package{}, err
	}
	return makePackage(path)
}

// Check if given path leads to an existing directory
func validateIsDir(path string) error {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return errors.New("path doesn't exist or not accessible")
	}
	if !fileInfo.IsDir() {
		return errors.New("not a directory")
	}
	return nil
}

// Check if directory at given path is an AUR package
func validateIsPackage(path string) error {
	if _, err := os.Stat(pkgbuildPath(path)); os.IsNotExist(err) {
		return errors.New("not a package: missing PKGBUILD")
	}
	if _, err := os.Stat(srcinfoPath(path)); os.IsNotExist(err) {
		return errors.New("not a package: missing .SRCINFO")
	}
	return nil
}

// Get path to PKGBUILD given package root path
func pkgbuildPath(path string) string {
	return filepath.Join(path, "PKGBUILD")
}

// Get path to .SRCINFO given package root path
func srcinfoPath(path string) string {
	return filepath.Join(path, ".SRCINFO")
}

// Create `Package` struct based on given package path dir without any safety checks
func makePackage(path string) (*Package, error) {
	srcinfoBytes, err := os.ReadFile(srcinfoPath(path))
	if err != nil {
		return &Package{}, err
	}
	srcinfoText := string(srcinfoBytes[:])
	srcinfoReader := strings.NewReader(srcinfoText)

	pkgname, err := readSrcinfoField(srcinfoReader, "pkgname")
	if err != nil {
		return &Package{}, err
	}
	pkgver, err := readSrcinfoField(srcinfoReader, "pkgver")
	if err != nil {
		return &Package{}, err
	}
	pkgrel, err := readSrcinfoField(srcinfoReader, "pkgrel")
	if err != nil {
		return &Package{}, err
	}
	url, err := readSrcinfoField(srcinfoReader, "url")
	if err != nil {
		return &Package{}, err
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return &Package{}, err
	}

	return &Package{
		Path:    absPath,
		Pkgname: pkgname,
		Url:     url,
		Pkgver:  pkgver,
		Pkgrel:  pkgrel,
	}, nil
}

// Get the value of a field with the given name from .SRCINFO contents reader
func readSrcinfoField(srcinfo io.ReadSeeker, field string) (string, error) {
	srcinfo.Seek(0, 0)
	scanner := bufio.NewScanner(srcinfo)
	searchToken := strings.Join([]string{field, srcinfoSeparator}, "")
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, searchToken) {
			split := strings.SplitN(line, srcinfoSeparator, 2)
			return split[1], nil
		}
	}
	return "", fmt.Errorf("SRCINFO missing '%s' field", field)
}
