package testutils

import (
	"os"
	"path/filepath"
)

func CreatePackage(path string, pkgbuild []byte, srcinfo []byte) error {
	err := os.MkdirAll(path, 0o755)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(path, "PKGBUILD"), pkgbuild, 0o644)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(path, ".SRCINFO"), srcinfo, 0o644)
	if err != nil {
		return err
	}
	return nil
}
