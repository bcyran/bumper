package testutils

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func CreatePackage(path string, pkgbuild []byte, srcinfo []byte) {
	os.MkdirAll(path, 0o755)
	ioutil.WriteFile(filepath.Join(path, "PKGBUILD"), pkgbuild, 0o644)
	ioutil.WriteFile(filepath.Join(path, ".SRCINFO"), srcinfo, 0o644)
}
