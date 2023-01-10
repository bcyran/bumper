package bumper

import (
	"fmt"

	"github.com/bcyran/bumper/pack"
	"github.com/bcyran/bumper/upstream"
)

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

	provider := action.versionProviderFactory(pkg.Url)
	if provider == nil {
		actionResult.Status = ACTION_FAILED
		return actionResult
	}

	upstreamVersion, err := provider.LatestVersion()
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
