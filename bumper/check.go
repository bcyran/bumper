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

func (result checkActionResult) String() string {
	if result.Status == ACTION_FAILED {
		return "x"
	}
	if result.Status == ACTION_SKIPPED {
		return "-"
	}
	if result.cmpResult == 1 {
		return fmt.Sprintf("%s -> %s", result.currentVersion, result.upstreamVersion)
	} else if result.cmpResult == 0 {
		return fmt.Sprintf("%s ✓", result.currentVersion)
	} else {
		return fmt.Sprintf("%s < %s !", result.currentVersion, result.upstreamVersion)
	}
}

type versionProviderFactory func(string) upstream.VersionProvider

type CheckAction struct {
	versionProviderFactory versionProviderFactory
}

func NewCheckAction(versionProviderFactory versionProviderFactory) *CheckAction {
	return &CheckAction{versionProviderFactory: versionProviderFactory}
}

func (action *CheckAction) Execute(pkg *pack.Package) *checkActionResult {
	provider := action.versionProviderFactory(pkg.Url)
	if provider == nil {
		return &checkActionResult{BaseActionResult: BaseActionResult{Status: ACTION_FAILED}}
	}
	upstreamVersion, err := provider.LatestVersion()
	if err != nil {
		return &checkActionResult{BaseActionResult: BaseActionResult{Status: ACTION_FAILED}}
	}
	cmpResult := pack.VersionCmp(upstreamVersion, pkg.Pkgver)
	pkg.UpstreamVersion = upstreamVersion
	return &checkActionResult{
		BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS},
		currentVersion:   pkg.Pkgver,
		upstreamVersion:  upstreamVersion,
		cmpResult:        cmpResult,
	}
}
