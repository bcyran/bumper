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

	if err := action.bump(pkg); err != nil {
		actionResult.Status = ACTION_FAILED
		actionResult.Error = err
		actionResult.bumpOk = false
		return actionResult
	} else {
		actionResult.bumpOk = true
	}

	if err := action.updpkgsums(pkg); err != nil {
		actionResult.Status = ACTION_FAILED
		actionResult.Error = err
		actionResult.updpkgsumsOk = false
		return actionResult
	} else {
		actionResult.updpkgsumsOk = true
	}

	if err := action.makepkg(pkg); err != nil {
		actionResult.Status = ACTION_FAILED
		actionResult.Error = err
		actionResult.makepkgOk = false
		return actionResult
	} else {
		actionResult.makepkgOk = true
	}

	actionResult.Status = ACTION_SUCCESS

	return actionResult
}

func (action *BumpAction) bump(pkg *pack.Package) error {
	pkgbuild, err := os.ReadFile(pkg.PkgbuildPath())
	if err != nil {
		return err
	}
	updatedPkgbuild := strings.ReplaceAll(
		string(pkgbuild), pkg.Pkgver.GetVersionStr(), pkg.UpstreamVersion.GetVersionStr(),
	)
	if pkg.Pkgrel != "1" {
		updatedPkgbuild = pkgrelPattern.ReplaceAllString(updatedPkgbuild, newPkgrel)
	}
	err = os.WriteFile(pkg.PkgbuildPath(), []byte(updatedPkgbuild), 0644)
	if err != nil {
		return err
	}
	return nil
}

func (action *BumpAction) updpkgsums(pkg *pack.Package) error {
	_, err := action.commandRunner(pkg.Path, "updpkgsums")
	return err
}

func (action *BumpAction) makepkg(pkg *pack.Package) error {
	srcinfo, err := action.commandRunner(pkg.Path, "makepkg", "--printsrcinfo")
	if err != nil {
		return err
	}
	return os.WriteFile(pkg.SrcinfoPath(), srcinfo, 0644)
}
