package bumper

import (
	"os"
	"path/filepath"
)

// CollectPackages recursively finds AUR packages in the given directory,
// up to a given depth.
// Depth 0 means only given directory is searched.
// Depth 1 checks subdirectories as well, etc...
func CollectPackages(path string, depth int) ([]Package, error) {
	if err := validateIsDir(path); err != nil {
		return []Package{}, err
	}

	return collectPackages(path, depth), nil
}

func collectPackages(path string, depth int) []Package {
	packages := []Package{}
	if fileInfo, err := os.Stat(path); err != nil || !fileInfo.IsDir() {
		return packages
	}
	if pack, err := LoadPackage(path); err == nil {
		return []Package{*pack}
	}
	if depth <= 0 {
		return packages
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return packages
	}
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name()[0] == '.' {
			continue
		}
		entryPath := filepath.Join(path, entry.Name())
		entryPackages := collectPackages(entryPath, depth-1)
		packages = append(packages, entryPackages...)
	}
	return packages
}
