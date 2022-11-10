package bumper

import (
	"sync"

	"github.com/bcyran/bumper/upstream"
)

// CheckPackages tries to fetch the latest upstream version for slice of packages.
func CheckPackages(packages []Package) {
	var wg sync.WaitGroup
	for idx := range packages {
		wg.Add(1)
		go checkPackage(&packages[idx], &wg)
	}
	wg.Wait()
}

func checkPackage(pack *Package, wg *sync.WaitGroup) {
	defer wg.Done()
	provider := upstream.NewVersionProvider(pack.Url)
	if provider != nil {
		upstreamVersion, err := provider.LatestVersion()
		if err == nil {
			pack.UpstreamVersion = upstreamVersion
		}
	}
}
