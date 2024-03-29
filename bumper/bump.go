package bumper

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/bcyran/bumper/pack"
)

const newPkgrel = "pkgrel=1"

var (
	pkgrelPattern = regexp.MustCompile(`pkgrel=\d`)
	ErrBumpAction = errors.New("bump action error")
)

type bumpActionResult struct {
	BaseActionResult
	bumpOk       bool
	updpkgsumsOk bool
	makepkgOk    bool
}

func (result *bumpActionResult) String() string {
	if result.Status == ActionSkippedStatus {
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
		actionResult.Status = ActionSkippedStatus
		return actionResult
	}

	if err := action.bump(pkg); err != nil {
		actionResult.Status = ActionFailedStatus
		actionResult.Error = fmt.Errorf("%w: %w", ErrBumpAction, err)
		actionResult.bumpOk = false
		return actionResult
	}
	actionResult.bumpOk = true

	if err := action.updpkgsums(pkg); err != nil {
		actionResult.Status = ActionFailedStatus
		actionResult.Error = fmt.Errorf("%w: %w", ErrBumpAction, err)
		actionResult.updpkgsumsOk = false
		return actionResult
	}
	actionResult.updpkgsumsOk = true

	if err := action.makepkg(pkg); err != nil {
		actionResult.Status = ActionFailedStatus
		actionResult.Error = fmt.Errorf("%w: %w", ErrBumpAction, err)
		actionResult.makepkgOk = false
		return actionResult
	}
	actionResult.makepkgOk = true

	actionResult.Status = ActionSuccessStatus

	return actionResult
}

func (action *BumpAction) bump(pkg *pack.Package) error {
	pkgbuild, err := os.ReadFile(pkg.PkgbuildPath())
	if err != nil {
		return fmt.Errorf("PKGBUILD reading error: %w", err)
	}
	updatedPkgbuild := strings.ReplaceAll(
		string(pkgbuild), pkg.Pkgver.GetVersionStr(), pkg.UpstreamVersion.GetVersionStr(),
	)
	if pkg.Pkgrel != "1" {
		updatedPkgbuild = pkgrelPattern.ReplaceAllString(updatedPkgbuild, newPkgrel)
	}
	err = os.WriteFile(pkg.PkgbuildPath(), []byte(updatedPkgbuild), 0o644)
	if err != nil {
		return fmt.Errorf("PKGBUILD writing error: %w", err)
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
	err = os.WriteFile(pkg.SrcinfoPath(), srcinfo, 0o644)
	if err != nil {
		return fmt.Errorf(".SRCINFO writing error: %w", err)
	}
	return nil
}
