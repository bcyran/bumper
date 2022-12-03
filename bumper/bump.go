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
		return "bump failed"
	}
	if !result.updpkgsumsOk {
		return "updpkgsums failed"
	}
	if !result.makepkgOk {
		return "makepkg failed"
	}
	return "bumped"
}

type BumpAction struct {
	commandRunner CommandRunner
}

func NewBumpAction(commandRunner CommandRunner) *BumpAction {
	return &BumpAction{commandRunner: commandRunner}
}

func (action *BumpAction) Execute(pkg *pack.Package) ActionResult {
	actionResult := &bumpActionResult{}

	if !pkg.IsOutdated {
		actionResult.Status = ACTION_SKIPPED
		return actionResult
	}

	if result := action.bump(pkg); result == false {
		actionResult.Status = ACTION_FAILED
		actionResult.bumpOk = false
		return actionResult
	} else {
		actionResult.bumpOk = true
	}

	if result := action.updpkgsums(pkg); result == false {
		actionResult.Status = ACTION_FAILED
		actionResult.updpkgsumsOk = false
		return actionResult
	} else {
		actionResult.updpkgsumsOk = true
	}

	if result := action.makepkg(pkg); result == false {
		actionResult.Status = ACTION_FAILED
		actionResult.makepkgOk = false
		return actionResult
	} else {
		actionResult.makepkgOk = true
	}

	actionResult.Status = ACTION_SUCCESS

	return actionResult
}

func (action *BumpAction) bump(pkg *pack.Package) bool {
	pkgbuild, err := os.ReadFile(pkg.PkgbuildPath())
	if err != nil {
		return false
	}
	updatedPkgbuild := strings.ReplaceAll(
		string(pkgbuild), pkg.Pkgver.GetVersionStr(), pkg.UpstreamVersion.GetVersionStr(),
	)
	if pkg.Pkgrel != "1" {
		updatedPkgbuild = pkgrelPattern.ReplaceAllString(updatedPkgbuild, newPkgrel)
	}
	err = os.WriteFile(pkg.PkgbuildPath(), []byte(updatedPkgbuild), 0644)
	if err != nil {
		return false
	}
	return true
}

func (action *BumpAction) updpkgsums(pkg *pack.Package) bool {
	_, err := action.commandRunner(pkg.Path, "updpkgsums")
	if err != nil {
		return false
	}
	return true
}

func (action *BumpAction) makepkg(pkg *pack.Package) bool {
	srcinfo, err := action.commandRunner(pkg.Path, "makepkg", "--printsrcinfo")
	if err != nil {
		return false
	}
	err = os.WriteFile(pkg.SrcinfoPath(), srcinfo, 0644)
	if err != nil {
		return false
	}
	return true
}
