package bumper

import (
	"errors"
	"fmt"

	"github.com/bcyran/bumper/pack"
)

var makeError = errors.New("make action error")

type makeActionResult struct {
	BaseActionResult
}

func (result *makeActionResult) String() string {
	if result.Status == ACTION_SKIPPED {
		return ""
	}
	if result.Status == ACTION_FAILED {
		return "build failed"
	}
	return "built"
}

type MakeAction struct {
	commandRunner CommandRunner
}

func NewMakeAction(commandRunner CommandRunner) *MakeAction {
	return &MakeAction{commandRunner: commandRunner}
}

func (action *MakeAction) Execute(pkg *pack.Package) ActionResult {
	actionResult := &makeActionResult{}

	if !pkg.IsOutdated {
		actionResult.Status = ACTION_SKIPPED
		return actionResult
	}

	_, err := action.commandRunner(pkg.Path, "makepkg", "--force", "--clean")
	if err != nil {
		actionResult.Status = ACTION_FAILED
		actionResult.Error = fmt.Errorf("%w: %w", makeError, err)
	} else {
		actionResult.Status = ACTION_SUCCESS
	}

	return actionResult
}
