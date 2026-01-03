package bumper

import (
	"os"
	"path/filepath"

	"github.com/bcyran/bumper/pack"
)

// CollectPackages recursively finds AUR packages in the given directory,
// up to a given depth.
// Depth 0 means only given directory is searched.
// Depth 1 checks subdirectories as well, etc...
func CollectPackages(path string, depth int) ([]pack.Package, error) {
	if err := pack.ValidateIsDir(path); err != nil {
		return []pack.Package{}, err
	}

	return collectPackages(path, depth), nil
}

func collectPackages(path string, depth int) []pack.Package {
	DebugLogger.Printf("Checking directory: %s (depth: %d)", path, depth)

	packages := []pack.Package{}
	if fileInfo, err := os.Stat(path); err != nil || !fileInfo.IsDir() {
		if err != nil {
			DebugLogger.Printf("  -> Skipped: %s (error: %v)", path, err)
		} else {
			DebugLogger.Printf("  -> Skipped: %s (not a directory)", path)
		}
		return packages
	}

	if pkg, err := pack.LoadPackage(path); err == nil {
		DebugLogger.Printf("  -> Found AUR package: %s", pkg.Pkgbase)
		return []pack.Package{*pkg}
	} else {
		DebugLogger.Printf("  -> Not a package: %s (%v)", path, err)
	}

	if depth <= 0 {
		DebugLogger.Printf("  -> Reached maximum depth at: %s", path)
		return packages
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		DebugLogger.Printf("  -> Cannot read directory: %s (error: %v)", path, err)
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
