package bumper

import (
	"fmt"
	"strings"

	"github.com/bcyran/bumper/pack"
	"github.com/bcyran/bumper/upstream"
)

const sourceSeparator = "::"

type checkActionResult struct {
	BaseActionResult
	currentVersion  pack.Version
	upstreamVersion upstream.Version
	cmpResult       int
}

func (result *checkActionResult) String() string {
	if result.Status == ACTION_FAILED {
		return "?"
	}
	if result.Status == ACTION_SKIPPED {
		return "-"
	}
	if result.cmpResult == 1 {
		return fmt.Sprintf("%s -> %s", result.currentVersion, result.upstreamVersion)
	} else if result.cmpResult == 0 {
		return fmt.Sprintf("%s", result.currentVersion)
	} else {
		return fmt.Sprintf("%s < %s !", result.upstreamVersion, result.currentVersion)
	}
}

type versionProviderFactory func(string) upstream.VersionProvider

type CheckAction struct {
	versionProviderFactory versionProviderFactory
}

func NewCheckAction(versionProviderFactory versionProviderFactory) *CheckAction {
	return &CheckAction{versionProviderFactory: versionProviderFactory}
}

func (action *CheckAction) Execute(pkg *pack.Package) ActionResult {
	actionResult := &checkActionResult{}

	if pkg.IsVCS {
		actionResult.Status = ACTION_SKIPPED
		return actionResult
	}

	upstreamUrls := getPackageUrls(pkg)
	upstreamVersion, err := action.tryGetUpstreamVersion(upstreamUrls)
	if err != nil {
		actionResult.Status = ACTION_FAILED
		return actionResult
	}

	cmpResult := pack.VersionCmp(upstreamVersion, pkg.Pkgver)
	pkg.UpstreamVersion = upstreamVersion
	pkg.IsOutdated = cmpResult == 1

	actionResult.Status = ACTION_SUCCESS
	actionResult.currentVersion = pkg.Pkgver
	actionResult.upstreamVersion = upstreamVersion
	actionResult.cmpResult = cmpResult

	return actionResult
}

// tryGetUpstreamVersion tries to create and use a version provider for each of the given URLs.
func (action *CheckAction) tryGetUpstreamVersion(urls []string) (upstream.Version, error) {
	providers := []upstream.VersionProvider{}
	for _, url := range urls {
		if provider := action.versionProviderFactory(url); provider != nil {
			providers = append(providers, provider)
		}
	}

	if len(providers) == 0 {
		return upstream.Version(""), fmt.Errorf("no upstream provider found")
	}

	var upstreamErr error
	for _, provider := range providers {
		upstreamVersion, err := provider.LatestVersion()
		if err == nil {
			return upstreamVersion, nil
		} else {
			upstreamErr = err
		}
	}

	return upstream.Version(""), fmt.Errorf("upstream provider error: %w", upstreamErr)
}

// getPackageUrls extracts all relevant URLs from given package.
// This includes both 'url' field and 'source' fields.
func getPackageUrls(pkg *pack.Package) []string {
	var urls = []string{pkg.Url}

	for _, sourceEntry := range pkg.Source {
		// source entry might be in the form 'file.bar::https://path/to/file.bar'
		_, sourceUrl, separatorFound := strings.Cut(sourceEntry, sourceSeparator)
		if !separatorFound {
			sourceUrl = sourceEntry
		}
		urls = append(urls, sourceUrl)
	}

	return urls
}
