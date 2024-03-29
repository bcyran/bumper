package bumper

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bcyran/bumper/pack"
	"github.com/bcyran/bumper/upstream"
	"go.uber.org/config"
)

var ErrCheckAction = errors.New("check action error")

const sourceSeparator = "::"

type checkActionResult struct {
	BaseActionResult
	currentVersion  pack.Version
	upstreamVersion upstream.Version
	cmpResult       int
}

func (result *checkActionResult) String() string {
	if result.Status == ActionFailedStatus {
		return "?"
	}
	if result.Status == ActionSkippedStatus {
		return result.currentVersion.GetVersionStr()
	}
	if result.cmpResult == 1 {
		return fmt.Sprintf("%s → %s", result.currentVersion, result.upstreamVersion)
	} else if result.cmpResult == 0 {
		return result.currentVersion.GetVersionStr()
	} else { // nolint:revive
		return fmt.Sprintf("%s < %s !", result.upstreamVersion, result.currentVersion)
	}
}

type versionProviderFactory func(string, config.Value) upstream.VersionProvider

type CheckAction struct {
	versionProviderFactory versionProviderFactory
	checkConfig            config.Value
}

func NewCheckAction(versionProviderFactory versionProviderFactory, checkConfig config.Value) *CheckAction {
	return &CheckAction{versionProviderFactory: versionProviderFactory, checkConfig: checkConfig}
}

func (action *CheckAction) Execute(pkg *pack.Package) ActionResult {
	actionResult := &checkActionResult{}

	var upstreamVersion upstream.Version

	var pkgVersionOverride string
	action.checkConfig.Get("versionOverrides").Get(pkg.Pkgbase).Populate(&pkgVersionOverride) // nolint:errcheck

	if pkgVersionOverride != "" {
		var isValid bool
		upstreamVersion, isValid = upstream.ParseVersion(pkgVersionOverride)
		if !isValid {
			actionResult.Status = ActionFailedStatus
			actionResult.Error = fmt.Errorf("%w: version override '%s' is not a valid version", ErrCheckAction, pkgVersionOverride)
			return actionResult
		}
	} else {
		if pkg.IsVCS {
			actionResult.currentVersion = pkg.Pkgver
			actionResult.Status = ActionSkippedStatus
			return actionResult
		}

		upstreamUrls := getPackageUrls(pkg)
		var err error
		upstreamVersion, err = action.tryGetUpstreamVersion(upstreamUrls)
		if err != nil {
			actionResult.Status = ActionFailedStatus
			actionResult.Error = err
			return actionResult
		}
	}

	cmpResult := pack.VersionCmp(upstreamVersion, pkg.Pkgver)
	pkg.UpstreamVersion = upstreamVersion
	pkg.IsOutdated = cmpResult == 1

	actionResult.Status = ActionSuccessStatus
	actionResult.currentVersion = pkg.Pkgver
	actionResult.upstreamVersion = upstreamVersion
	actionResult.cmpResult = cmpResult

	return actionResult
}

// getPackageUrls extracts all relevant URLs from given package.
// This includes both 'url' field and 'source' fields.
func getPackageUrls(pkg *pack.Package) []string {
	urls := []string{pkg.URL}

	for _, sourceEntry := range pkg.Source {
		// source entry might be in the form 'file.bar::https://path/to/file.bar'
		_, sourceURL, separatorFound := strings.Cut(sourceEntry, sourceSeparator)
		if !separatorFound {
			sourceURL = sourceEntry
		}
		urls = append(urls, sourceURL)
	}

	return urls
}

// tryGetUpstreamVersion tries to create and use a version provider for each of the given URLs.
func (action *CheckAction) tryGetUpstreamVersion(urls []string) (upstream.Version, error) {
	providersConfig := action.checkConfig.Get("providers")
	providers := []upstream.VersionProvider{}
	for _, url := range urls {
		if newProvider := action.versionProviderFactory(url, providersConfig); newProvider != nil {
			providers = appendUnique(providers, newProvider)
		}
	}

	if len(providers) == 0 {
		return upstream.Version(""), fmt.Errorf("no upstream provider found")
	}

	upstreamErrs := []error{}
	for _, provider := range providers {
		upstreamVersion, err := provider.LatestVersion()
		if err == nil {
			return upstreamVersion, nil
		}
		upstreamErrs = append(upstreamErrs, fmt.Errorf("upstream provider error: %w", err))
	}

	return upstream.Version(""), errors.Join(upstreamErrs...)
}

func appendUnique(providers []upstream.VersionProvider, newProvider upstream.VersionProvider) []upstream.VersionProvider {
	isUnique := true
	for _, existingProvider := range providers {
		if newProvider.Equal(existingProvider) {
			isUnique = false
		}
	}
	if isUnique {
		providers = append(providers, newProvider)
	}
	return providers
}
