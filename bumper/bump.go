package bumper

import (
	"os"
	"regexp"
	"strings"

	"github.com/bcyran/bumper/pack"
)

const newPkgrel = "pkgrel=1"

var pkgrelPattern = regexp.MustCompile(`pkgrel=\d`)

type bumpActionResult struct {
	BaseActionResult
	bumpOk       bool
	updpkgsumsOk bool
	makepkgOk    bool
}

func (result *bumpActionResult) String() string {
	if result.Status == ACTION_SKIPPED {
		return ""
	}
	if !result.bumpOk {
		return "bump ✗"
	}
	if !result.updpkgsumsOk {
		return "updpkgsums ✗"
	}
	if !result.makepkgOk {
		return "makepkg ✗"
	}
	return "bump ✓"
}

type BumpAction struct{}

func (action *BumpAction) Execute(pkg *pack.Package) *bumpActionResult {
	if !pkg.IsOutdated {
		return &bumpActionResult{BaseActionResult: BaseActionResult{Status: ACTION_SKIPPED}}
	}
	if result := action.bump(pkg); result != nil {
		return result
	}

	return &bumpActionResult{
		BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS},
		bumpOk:           true,
		updpkgsumsOk:     true,
		makepkgOk:        true,
	}
}

func (action *BumpAction) bump(pkg *pack.Package) *bumpActionResult {
	pkgbuild, err := os.ReadFile(pkg.PkgbuildPath())
	if err != nil {
		return &bumpActionResult{BaseActionResult: BaseActionResult{Status: ACTION_FAILED}, bumpOk: false}
	}
	updatedPkgbuild := strings.ReplaceAll(
		string(pkgbuild), pkg.Pkgver.GetVersionStr(), pkg.UpstreamVersion.GetVersionStr(),
	)
	if pkg.Pkgrel != "1" {
		updatedPkgbuild = pkgrelPattern.ReplaceAllString(updatedPkgbuild, newPkgrel)
	}
	err = os.WriteFile(pkg.PkgbuildPath(), []byte(updatedPkgbuild), 0644)
	if err != nil {
		return &bumpActionResult{BaseActionResult: BaseActionResult{Status: ACTION_FAILED}, bumpOk: false}
	}
	return nil
}

func (action *BumpAction) updpkgsums(pkg *pack.Package) *bumpActionResult {
	return nil
}

func (action *BumpAction) makepkg(pkg *pack.Package) *bumpActionResult {
	return nil
}
